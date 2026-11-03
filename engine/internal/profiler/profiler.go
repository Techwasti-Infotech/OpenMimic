package profiler

import (
	"fmt"
	"math"
	"sort"
	"strings"

	pb "openmimic/engine/pkg/proto/v1"
)

// ColumnProfile captures statistical inferences from sampled database records.
type ColumnProfile struct {
	Name         string
	InferredType string // "uuid", "numeric", "categorical", "timestamp", "string"
	NumericMean  float64
	NumericStd   float64
	Categories   map[string]int
	IsPrimaryKey bool
}

// DatabaseProfiler reverse-engineers generator configurations from sample database rows.
type DatabaseProfiler struct {
	tableName string
	rows      []map[string]any
}

func NewDatabaseProfiler(tableName string, rows []map[string]any) *DatabaseProfiler {
	return &DatabaseProfiler{
		tableName: tableName,
		rows:      rows,
	}
}

// Profile analyzes samples and constructs a high-fidelity GeneratorConfig.
func (p *DatabaseProfiler) Profile() (*pb.GeneratorConfig, error) {
	if len(p.rows) == 0 {
		return nil, fmt.Errorf("no sample rows provided for profiling")
	}

	// 1. Collect all columns
	colNames := make([]string, 0)
	colSeen := make(map[string]bool)
	for _, row := range p.rows {
		for k := range row {
			if !colSeen[k] {
				colSeen[k] = true
				colNames = append(colNames, k)
			}
		}
	}
	sort.Strings(colNames)

	fields := make([]*pb.Field, 0, len(colNames))

	for _, col := range colNames {
		gen := p.profileColumn(col)
		fields = append(fields, &pb.Field{
			Name:      col,
			Generator: gen,
		})
	}

	return &pb.GeneratorConfig{
		Id:                 fmt.Sprintf("profiled_%s", p.tableName),
		Name:               fmt.Sprintf("Profiled Stream: %s", p.tableName),
		Seed:               42,
		EventLimit:         1000,
		RateLimitPerSecond: 25,
		Sink: &pb.SinkConfig{
			SinkType: &pb.SinkConfig_Kafka{
				Kafka: &pb.KafkaSink{
					Brokers: []string{"localhost:9092"},
					Topic:   fmt.Sprintf("%s.events", p.tableName),
				},
			},
		},
		Fields: fields,
	}, nil
}

func (p *DatabaseProfiler) profileColumn(col string) *pb.FunctionGenerator {
	lower := strings.ToLower(col)

	// Check for UUID patterns
	if strings.Contains(lower, "id") || strings.Contains(lower, "uuid") {
		return &pb.FunctionGenerator{
			Func: &pb.FunctionGenerator_Uuid{
				Uuid: &pb.UuidFunction{Uppercase: false},
			},
		}
	}

	// Check for timestamps (only exact suffix or dedicated word like created_at, updated_at, timestamp, datetime)
	if lower == "timestamp" || lower == "datetime" || strings.HasSuffix(lower, "_at") || lower == "date" {
		return &pb.FunctionGenerator{
			Func: &pb.FunctionGenerator_Now{
				Now: &pb.NowFunction{Format: "RFC3339"},
			},
		}
	}

	// Inspect sampled values
	var numCount int
	var numValues []float64
	categories := make(map[string]int)

	for _, row := range p.rows {
		val := row[col]
		if val == nil {
			continue
		}

		switch v := val.(type) {
		case float64:
			numCount++
			numValues = append(numValues, v)
		case int:
			numCount++
			numValues = append(numValues, float64(v))
		case int64:
			numCount++
			numValues = append(numValues, float64(v))
		case string:
			categories[v]++
		}
	}

	// Numeric distribution inference
	if numCount > 0 && numCount >= len(p.rows)/2 {
		var sum float64
		for _, v := range numValues {
			sum += v
		}
		mean := sum / float64(len(numValues))

		var variance float64
		for _, v := range numValues {
			variance += math.Pow(v-mean, 2)
		}
		stdDev := math.Sqrt(variance / float64(len(numValues)))
		if stdDev == 0 {
			stdDev = 1.0
		}

		return &pb.FunctionGenerator{
			Func: &pb.FunctionGenerator_Distribution{
				Distribution: &pb.DistributionFunction{
					DistributionType: &pb.DistributionFunction_Normal{
						Normal: &pb.NormalDistribution{
							Mean:   math.Round(mean*100) / 100,
							StdDev: math.Round(stdDev*100) / 100,
						},
					},
				},
			},
		}
	}

	// Categorical choice inference
	if len(categories) > 0 && len(categories) <= 10 {
		var total int
		for _, count := range categories {
			total += count
		}
		options := make([]*pb.WeightedChoiceOption, 0, len(categories))
		for cat, count := range categories {
			weight := (float64(count) / float64(total)) * 100.0
			options = append(options, &pb.WeightedChoiceOption{
				Value:  cat,
				Weight: math.Round(weight*10) / 10,
			})
		}
		return &pb.FunctionGenerator{
			Func: &pb.FunctionGenerator_WeightedChoice{
				WeightedChoice: &pb.WeightedChoiceFunction{Options: options},
			},
		}
	}

	// Fallback: contextual string faker
	cat := "name"
	if strings.Contains(lower, "mail") {
		cat = "email"
	} else if strings.Contains(lower, "city") {
		cat = "city"
	}
	return &pb.FunctionGenerator{
		Func: &pb.FunctionGenerator_StringFaker{
			StringFaker: &pb.StringFakerFunction{Category: cat},
		},
	}
}
