package telemetry

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

// PrometheusExporter formats live engine metrics into standard Prometheus text format.
type PrometheusExporter struct {
	mu           sync.RWMutex
	eventCounts  map[string]int64
	errorCounts  map[string]int64
	latencies    map[string]float64
	throughputs  map[string]float64
}

var (
	defaultExporter     *PrometheusExporter
	defaultExporterOnce sync.Once
)

func GetDefaultExporter() *PrometheusExporter {
	defaultExporterOnce.Do(func() {
		defaultExporter = &PrometheusExporter{
			eventCounts: make(map[string]int64),
			errorCounts: make(map[string]int64),
			latencies:   make(map[string]float64),
			throughputs: make(map[string]float64),
		}
	})
	return defaultExporter
}

func (e *PrometheusExporter) Record(genID, sinkType string, events, errors int64, latencyP99, eps float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := fmt.Sprintf("%s_%s", genID, sinkType)
	e.eventCounts[key] = events
	e.errorCounts[key] = errors
	e.latencies[key] = latencyP99
	e.throughputs[key] = eps
}

// Scrape returns Prometheus text format (version 0.0.4).
func (e *PrometheusExporter) Scrape() string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var b strings.Builder
	b.WriteString("# HELP openmimic_events_generated_total Total synthetic records generated\n")
	b.WriteString("# TYPE openmimic_events_generated_total counter\n")
	for k, count := range e.eventCounts {
		parts := strings.SplitN(k, "_", 2)
		b.WriteString(fmt.Sprintf("openmimic_events_generated_total{generator_id=\"%s\",sink=\"%s\"} %d\n", parts[0], parts[1], count))
	}

	b.WriteString("\n# HELP openmimic_errors_total Total errors and infeasible retries\n")
	b.WriteString("# TYPE openmimic_errors_total counter\n")
	for k, errs := range e.errorCounts {
		parts := strings.SplitN(k, "_", 2)
		b.WriteString(fmt.Sprintf("openmimic_errors_total{generator_id=\"%s\",sink=\"%s\"} %d\n", parts[0], parts[1], errs))
	}

	b.WriteString("\n# HELP openmimic_throughput_eps Current generation rate (events/sec)\n")
	b.WriteString("# TYPE openmimic_throughput_eps gauge\n")
	for k, eps := range e.throughputs {
		parts := strings.SplitN(k, "_", 2)
		b.WriteString(fmt.Sprintf("openmimic_throughput_eps{generator_id=\"%s\",sink=\"%s\"} %.2f\n", parts[0], parts[1], eps))
	}

	b.WriteString("\n# HELP openmimic_latency_p99_milliseconds Event serialization and sink write latency p99\n")
	b.WriteString("# TYPE openmimic_latency_p99_milliseconds gauge\n")
	for k, p99 := range e.latencies {
		parts := strings.SplitN(k, "_", 2)
		b.WriteString(fmt.Sprintf("openmimic_latency_p99_milliseconds{generator_id=\"%s\",sink=\"%s\"} %.3f\n", parts[0], parts[1], p99))
	}

	return b.String()
}

// Handler returns an http.HandlerFunc that serves standard Prometheus scraping.
func (e *PrometheusExporter) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = w.Write([]byte(e.Scrape()))
	}
}

// OTelInjector injects W3C Trace Context traceparent headers into synthetic events.
type OTelInjector struct {
	cfg *pb.OTelTracingConfig
}

func NewOTelInjector(cfg *pb.OTelTracingConfig) *OTelInjector {
	return &OTelInjector{cfg: cfg}
}

// Inject returns a standard W3C traceparent (version 00, 16-byte traceId, 8-byte parentId, 1-byte traceFlags).
func (o *OTelInjector) Inject(record map[string]any, rng *state.DeterministicRand) map[string]any {
	if o.cfg == nil || !o.cfg.GetEnabled() {
		return record
	}

	rate := o.cfg.GetSampleRate()
	if rate <= 0 {
		rate = 1.0
	}

	if rng.Float64() <= rate {
		traceID := fmt.Sprintf("%016x%016x", rng.Uint64(), rng.Uint64())
		spanID := fmt.Sprintf("%016x", rng.Uint64())
		// W3C Trace Context: 00-{trace_id}-{span_id}-01 (sampled)
		traceparent := fmt.Sprintf("00-%s-%s-01", traceID, spanID)

		record["_traceparent"] = traceparent
		record["_otel_service"] = o.cfg.GetServiceName()
		record["_trace_sampled"] = true
	}

	return record
}
