package importer

import (
	"fmt"
	"regexp"
	"strings"

	pb "openmimic/engine/pkg/proto/v1"
)

// SQLDDLToGeneratorConfig parses standard SQL CREATE TABLE statements into a complete GeneratorConfig.
func SQLDDLToGeneratorConfig(ddl string) (*pb.GeneratorConfig, error) {
	// Extract table name: CREATE TABLE [IF NOT EXISTS] tableName (...)
	tableRegex := regexp.MustCompile(`(?i)create\s+table\s+(?:if\s+not\s+exists\s+)?([` + "`" + `"\[]?([a-zA-Z0-9_]+)[` + "`" + `"\]]?)`)
	tableMatch := tableRegex.FindStringSubmatch(ddl)
	tableName := "synthetic_table"
	if len(tableMatch) > 2 && tableMatch[2] != "" {
		tableName = tableMatch[2]
	}

	// Extract body inside parentheses
	firstParen := strings.Index(ddl, "(")
	lastParen := strings.LastIndex(ddl, ")")
	if firstParen == -1 || lastParen == -1 || lastParen <= firstParen {
		return nil, fmt.Errorf("unable to locate column definitions inside parentheses")
	}

	body := ddl[firstParen+1 : lastParen]
	lines := strings.Split(body, "\n")
	if len(lines) <= 1 {
		// Try splitting by comma
		lines = strings.Split(body, ",")
	}

	var fields []*pb.Field
	colRegex := regexp.MustCompile(`^\s*[` + "`" + `"\[]?([a-zA-Z0-9_]+)[` + "`" + `"\]]?\s+([a-zA-Z0-9_()]+)`)

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		line = strings.TrimSuffix(line, ",")
		if line == "" || strings.HasPrefix(strings.ToUpper(line), "PRIMARY KEY") ||
			strings.HasPrefix(strings.ToUpper(line), "CONSTRAINT") ||
			strings.HasPrefix(strings.ToUpper(line), "FOREIGN KEY") ||
			strings.HasPrefix(strings.ToUpper(line), "INDEX") ||
			strings.HasPrefix(strings.ToUpper(line), "UNIQUE") {
			continue
		}

		match := colRegex.FindStringSubmatch(line)
		if len(match) < 3 {
			continue
		}

		colName := match[1]
		dataType := strings.ToUpper(match[2])

		gen := inferGenerator(colName, dataType)
		fields = append(fields, &pb.Field{
			Name:      colName,
			Generator: gen,
		})
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("no valid columns detected in DDL statement")
	}

	cfg := &pb.GeneratorConfig{
		Id:                 fmt.Sprintf("gen_%s", tableName),
		Name:               fmt.Sprintf("Table: %s", tableName),
		Seed:               42,
		EventLimit:         1000,
		RateLimitPerSecond: 20,
		Sink: &pb.SinkConfig{
			SinkType: &pb.SinkConfig_Kafka{
				Kafka: &pb.KafkaSink{
					Brokers:  []string{"localhost:9092"},
					Topic:    fmt.Sprintf("db.%s", tableName),
					ClientId: "openmimic-importer",
				},
			},
		},
		Fields: fields,
	}

	return cfg, nil
}

func inferGenerator(colName, dataType string) *pb.FunctionGenerator {
	lowerCol := strings.ToLower(colName)

	// 1. UUID or ID columns
	if strings.HasSuffix(lowerCol, "_id") || lowerCol == "id" || strings.Contains(dataType, "UUID") {
		return &pb.FunctionGenerator{
			Func: &pb.FunctionGenerator_Uuid{
				Uuid: &pb.UuidFunction{Uppercase: false},
			},
		}
	}

	// 2. Timestamps / Dates
	if strings.Contains(dataType, "TIME") || strings.Contains(dataType, "DATE") {
		return &pb.FunctionGenerator{
			Func: &pb.FunctionGenerator_Now{
				Now: &pb.NowFunction{Format: "RFC3339"},
			},
		}
	}

	// 3. Status or Category codes
	if lowerCol == "status" || lowerCol == "state" {
		return &pb.FunctionGenerator{
			Func: &pb.FunctionGenerator_WeightedChoice{
				WeightedChoice: &pb.WeightedChoiceFunction{
					Options: []*pb.WeightedChoiceOption{
						{Value: "ACTIVE", Weight: 80.0},
						{Value: "PENDING", Weight: 15.0},
						{Value: "CANCELLED", Weight: 5.0},
					},
				},
			},
		}
	}

	// 4. Email addresses
	if strings.Contains(lowerCol, "email") {
		return &pb.FunctionGenerator{
			Func: &pb.FunctionGenerator_StringFaker{
				StringFaker: &pb.StringFakerFunction{Category: "email"},
			},
		}
	}

	// 5. Names
	if strings.Contains(lowerCol, "name") {
		return &pb.FunctionGenerator{
			Func: &pb.FunctionGenerator_StringFaker{
				StringFaker: &pb.StringFakerFunction{Category: "name"},
			},
		}
	}

	// 6. City / Locations
	if strings.Contains(lowerCol, "city") {
		return &pb.FunctionGenerator{
			Func: &pb.FunctionGenerator_StringFaker{
				StringFaker: &pb.StringFakerFunction{Category: "city"},
			},
		}
	}

	// 7. Numeric floating-point / monetary amounts
	if strings.Contains(dataType, "DECIMAL") || strings.Contains(dataType, "NUMERIC") ||
		strings.Contains(dataType, "FLOAT") || strings.Contains(dataType, "DOUBLE") ||
		strings.Contains(lowerCol, "amount") || strings.Contains(lowerCol, "price") ||
		strings.Contains(lowerCol, "balance") {
		return &pb.FunctionGenerator{
			Func: &pb.FunctionGenerator_Distribution{
				Distribution: &pb.DistributionFunction{
					DistributionType: &pb.DistributionFunction_Normal{
						Normal: &pb.NormalDistribution{Mean: 100.0, StdDev: 25.0},
					},
				},
			},
		}
	}

	// 8. Integer columns
	if strings.Contains(dataType, "INT") || strings.Contains(dataType, "SERIAL") {
		return &pb.FunctionGenerator{
			Func: &pb.FunctionGenerator_SequentialInteger{
				SequentialInteger: &pb.SequentialIntegerFunction{Start: 1, Step: 1},
			},
		}
	}

	// 9. Boolean columns
	if strings.Contains(dataType, "BOOL") {
		return &pb.FunctionGenerator{
			Func: &pb.FunctionGenerator_WeightedChoice{
				WeightedChoice: &pb.WeightedChoiceFunction{
					Options: []*pb.WeightedChoiceOption{
						{Value: "true", Weight: 90.0},
						{Value: "false", Weight: 10.0},
					},
				},
			},
		}
	}

	// Default fallback: pattern string
	return &pb.FunctionGenerator{
		Func: &pb.FunctionGenerator_StringFaker{
			StringFaker: &pb.StringFakerFunction{Category: "name"},
		},
	}
}

