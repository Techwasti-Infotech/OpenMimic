package engine

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"openmimic/engine/internal/cdc"
	"openmimic/engine/internal/dlq"
	"openmimic/engine/internal/generator"
	"openmimic/engine/internal/metrics"
	"openmimic/engine/internal/privacy"
	"openmimic/engine/internal/pseudonym"
	"openmimic/engine/internal/sink"
	"openmimic/engine/internal/solver"
	"openmimic/engine/internal/state"
	"openmimic/engine/internal/telemetry"
	"openmimic/engine/internal/traffic"
	pb "openmimic/engine/pkg/proto/v1"
)

// Engine coordinates event generation and writes to configured sinks.
type Engine struct {
	config     *pb.GeneratorConfig
	state      *state.GeneratorState
	sink       sink.Sink
	metrics    *metrics.Collector
	shaper     traffic.Shaper
	evaluator  *solver.ConstraintEvaluator
	injector   *generator.AnomalyInjector
	cdcEnv     *cdc.DebeziumEnvelope
	dlqRouter  *dlq.Router
	otelInj    *telemetry.OTelInjector
}

// NewEngine creates a new Engine instance.
func NewEngine(cfg *pb.GeneratorConfig, w io.Writer) *Engine {
	seed := cfg.GetSeed()
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	if w == nil {
		w = os.Stdout
	}

	st := state.NewGeneratorState(seed, time.Time{}, 100*time.Millisecond)

	s, err := sink.CreateSink(cfg.GetSink(), w)
	if err != nil {
		s = sink.NewStdoutSink(w, false)
	}

	shaper := traffic.NewShaper(cfg.GetTrafficProfile(), cfg.GetRateLimitPerSecond())
	evaluator := solver.NewConstraintEvaluator(cfg.GetConstraints())
	injector := generator.NewAnomalyInjector(cfg.GetPlaybooks())
	cdcEnv := cdc.NewDebeziumEnvelope(cfg.GetCdcEnvelope())
	dlqRouter := dlq.NewRouter(cfg.GetDlq())
	otelInj := telemetry.NewOTelInjector(cfg.GetOtel())

	return &Engine{
		config:    cfg,
		state:     st,
		sink:      s,
		metrics:   metrics.NewCollector(),
		shaper:    shaper,
		evaluator: evaluator,
		injector:  injector,
		cdcEnv:    cdcEnv,
		dlqRouter: dlqRouter,
		otelInj:   otelInj,
	}
}

// SetMetricsCollector allows wiring an external metrics collector (e.g. from gRPC server).
func (e *Engine) SetMetricsCollector(mc *metrics.Collector) {
	e.metrics = mc
}

// GetMetricsCollector returns the internal metrics collector.
func (e *Engine) GetMetricsCollector() *metrics.Collector {
	return e.metrics
}

// SetSink overrides the sink (used for testing or custom destinations).
func (e *Engine) SetSink(s sink.Sink) {
	e.sink = s
}

// GenerateEvent produces a single JSON map of field names to generated values satisfying constraints.
func (e *Engine) GenerateEvent(index int64) (map[string]any, error) {
	maxRetries := 100
	for _, c := range e.config.GetConstraints() {
		if c.GetMaxResampleRetries() > int32(maxRetries) {
			maxRetries = int(c.GetMaxResampleRetries())
		}
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		record := make(map[string]any)

		for _, field := range e.config.GetFields() {
			val := e.generateFieldValue(field.GetGenerator(), index)
			record[field.GetName()] = val
		}

		// Evaluate candidate against constraints
		satisfied, err := e.evaluator.Evaluate(record)
		if err != nil {
			return nil, err
		}
		if satisfied {
			// Apply anomaly injection if playbooks are active
			record, _ = e.injector.Apply(record, e.state.RNG)

			// Wrap in Debezium CDC envelope if enabled
			record = e.cdcEnv.Wrap(record, index, e.state.RNG, e.state.Clock)

			// Inject OpenTelemetry W3C traceparent headers if enabled
			record = e.otelInj.Inject(record, e.state.RNG)

			e.state.Clock.Tick()
			return record, nil
		}
	}

	err := fmt.Errorf("infeasible constraint set: unable to satisfy predicates after %d resample attempts", maxRetries)
	// Route failed candidate to Dead Letter Queue if configured
	_ = e.dlqRouter.Quarantine(context.Background(), e.config.GetId(), err.Error(), maxRetries, map[string]any{"index": index})

	return nil, err
}

func (e *Engine) generateFieldValue(gen *pb.FunctionGenerator, index int64) any {
	if gen == nil {
		return nil
	}

	switch f := gen.Func.(type) {
	case *pb.FunctionGenerator_Uuid:
		uppercase := false
		if f.Uuid != nil {
			uppercase = f.Uuid.Uppercase
		}
		return generator.GenerateUUID(e.state.RNG, uppercase)

	case *pb.FunctionGenerator_Distribution:
		if f.Distribution == nil {
			return 0.0
		}
		switch d := f.Distribution.DistributionType.(type) {
		case *pb.DistributionFunction_Normal:
			return generator.GenerateNormalDistribution(e.state.RNG, d.Normal.Mean, d.Normal.StdDev)
		case *pb.DistributionFunction_Uniform:
			return generator.GenerateUniformDistribution(e.state.RNG, d.Uniform.Min, d.Uniform.Max)
		}

	case *pb.FunctionGenerator_Now:
		format := "RFC3339"
		if f.Now != nil && f.Now.Format != "" {
			format = f.Now.Format
		}
		return generator.GenerateNow(e.state.Clock, format)

	case *pb.FunctionGenerator_SequentialInteger:
		start := int64(0)
		step := int64(1)
		if f.SequentialInteger != nil {
			start = f.SequentialInteger.Start
			if f.SequentialInteger.Step != 0 {
				step = f.SequentialInteger.Step
			}
		}
		return generator.GenerateSequentialInteger(index, start, step)

	case *pb.FunctionGenerator_StringFaker:
		cat := "word"
		pat := ""
		if f.StringFaker != nil {
			if f.StringFaker.Category != "" {
				cat = f.StringFaker.Category
			}
			pat = f.StringFaker.Pattern
		}
		return generator.GenerateStringFaker(e.state.RNG, cat, pat, index)

	case *pb.FunctionGenerator_Constant:
		if f.Constant != nil {
			switch val := f.Constant.Value.(type) {
			case *pb.ConstantFunction_StringVal:
				return val.StringVal
			case *pb.ConstantFunction_IntVal:
				return val.IntVal
			case *pb.ConstantFunction_DoubleVal:
				return val.DoubleVal
			case *pb.ConstantFunction_BoolVal:
				return val.BoolVal
			}
		}

	case *pb.FunctionGenerator_WeightedChoice:
		if f.WeightedChoice != nil {
			return generator.GenerateWeightedChoice(e.state.RNG, f.WeightedChoice.Options)
		}

	case *pb.FunctionGenerator_MarkovChain:
		if f.MarkovChain != nil {
			return generator.GenerateMarkovChain(e.state, fmt.Sprintf("field_state_%d", index), f.MarkovChain)
		}

	case *pb.FunctionGenerator_EntityPool:
		if f.EntityPool != nil {
			// Check if dynamic foreign key binding is active
			if fk := f.EntityPool.GetForeignKey(); fk != nil {
				targetPool := fk.GetParentPoolName()
				if targetPool == "" {
					targetPool = f.EntityPool.GetPoolName()
				}

				// Check deliberate orphan corruption injection
				if fk.GetAllowOrphans() && fk.GetOrphanRate() > 0 && e.state.RNG.Float64() < fk.GetOrphanRate() {
					return fmt.Sprintf("orphan_%08x", e.state.RNG.Uint32())
				}

				// Sample from live parent entity registry
				if entityID, found := state.GetDefaultRegistry().SampleEntity(targetPool, e.state.RNG); found {
					return entityID
				}
			}

			// Generate static deterministic entity pool ID
			entityID := generator.GenerateEntityPool(e.state.RNG, f.EntityPool)

			// Register into live entity registry so downstream generators can reference it
			if f.EntityPool.GetPoolName() != "" {
				state.GetDefaultRegistry().RegisterEntity(f.EntityPool.GetPoolName(), entityID)
			}
			return entityID
		}

	case *pb.FunctionGenerator_NestedObject:
		if f.NestedObject != nil {
			nestedRecord := make(map[string]any)
			for _, subField := range f.NestedObject.GetFields() {
				nestedRecord[subField.GetName()] = e.generateFieldValue(subField.GetGenerator(), index)
			}
			return nestedRecord
		}

	case *pb.FunctionGenerator_Array:
		if f.Array != nil {
			minItems := int(f.Array.GetMinItems())
			maxItems := int(f.Array.GetMaxItems())
			if maxItems < minItems {
				maxItems = minItems
			}
			count := minItems
			if maxItems > minItems {
				count = minItems + e.state.RNG.Intn(maxItems-minItems+1)
			}
			items := make([]any, 0, count)
			for j := 0; j < count; j++ {
				itemVal := e.generateFieldValue(f.Array.GetItemGenerator(), int64(j))
				items = append(items, itemVal)
			}
			return items
		}

	case *pb.FunctionGenerator_MediaMeta:
		if f.MediaMeta != nil {
			mediaType := f.MediaMeta.GetMediaType()
			if mediaType == "" {
				mediaType = "image"
			}
			mime := f.MediaMeta.GetMimeType()
			if mime == "" {
				mime = "image/jpeg"
			}
			w := f.MediaMeta.GetWidth()
			if w == 0 {
				w = 1920
			}
			h := f.MediaMeta.GetHeight()
			if h == 0 {
				h = 1080
			}
			size := f.MediaMeta.GetMaxBytes()
			if size == 0 {
				size = int64(1024 * 1024 * 2)
			}
			return map[string]any{
				"media_type":   mediaType,
				"mime_type":    mime,
				"width":        w,
				"height":       h,
				"size_bytes":   int64(float64(size) * (0.8 + e.state.RNG.Float64()*0.4)),
				"storage_uri":  fmt.Sprintf("s3://mimic-media/%s/%08x.%s", mediaType, e.state.RNG.Uint32(), strings.TrimPrefix(mime, mediaType+"/")),
				"checksum_sha": fmt.Sprintf("%016x%016x", e.state.RNG.Uint64(), e.state.RNG.Uint64()),
			}
		}

	case *pb.FunctionGenerator_DifferentialPrivacy:
		if f.DifferentialPrivacy != nil {
			dp, err := privacy.CreateDPMechanism(f.DifferentialPrivacy)
			if err == nil {
				baseVal := f.DifferentialPrivacy.GetBaseValue()
				return dp.AddNoise(baseVal, e.state.RNG)
			}
			return f.DifferentialPrivacy.GetBaseValue()
		}

	case *pb.FunctionGenerator_Pseudonymize:
		if f.Pseudonymize != nil {
			p := pseudonym.NewPseudonymizer(f.Pseudonymize)
			return p.Pseudonymize(e.state.RNG)
		}
	}

	return nil
}

// Run executes the generation loop writing events to the configured sink.
func (e *Engine) Run(ctx context.Context) error {
	defer e.sink.Close()

	limit := e.config.GetEventLimit()
	loopStart := time.Now()

	for i := int64(0); limit == 0 || i < limit; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			start := time.Now()
			record, err := e.GenerateEvent(i)
			if err != nil {
				e.metrics.RecordError()
				return err
			}

			key := []byte(fmt.Sprintf("%d", i))
			if err := e.sink.WriteEvent(ctx, key, record); err != nil {
				e.metrics.RecordError()
				return err
			}

			duration := time.Since(start)
			e.metrics.RecordEvent(duration)

			// Traffic profile delay calculation
			elapsed := time.Since(loopStart)
			delay := e.shaper.NextDelay(elapsed, e.state.RNG)
			if delay > 0 {
				time.Sleep(delay)
			}
		}
	}
	return nil
}
