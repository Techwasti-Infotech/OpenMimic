package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"openmimic/engine"
	"openmimic/engine/internal/cluster"
	"openmimic/engine/internal/importer"
	"openmimic/engine/internal/profiler"
	pb "openmimic/engine/pkg/proto/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func printUsage() {
	fmt.Println("Usage: mimic <command> [arguments]")
	fmt.Println("Commands:")
	fmt.Println("  validate --config <file> [--constraints]   Validate schema & predicates (returns non-zero on failure)")
	fmt.Println("  preview  --config <file> [--count 10]      Render sample events to stdout")
	fmt.Println("  run      --config <file> [--dry-run]       Execute generation pipeline")
	fmt.Println("  export   --to-template <name>              Generate a starter template (e.g. order-flow, user-signup)")
	fmt.Println("  discover --sql <file>                      Generate mimic.json from SQL DDL table definition")
	fmt.Println("  profile  --table <name> --sample <file>    Reverse-engineer mimic.json from sample JSON records")
	fmt.Println("  cluster  --workers <count> --total <count> Simulate distributed cluster partitioned stream")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "validate":
		validateCmd := flag.NewFlagSet("validate", flag.ExitOnError)
		configPath := validateCmd.String("config", "mimic.json", "Path to mimic.json")
		checkConstraints := validateCmd.Bool("constraints", true, "Check constraint feasibility")
		_ = validateCmd.Parse(os.Args[2:])

		cfg, err := loadConfig(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: schema validation failed: %v\n", err)
			os.Exit(1)
		}

		if *checkConstraints && len(cfg.GetConstraints()) > 0 {
			eng := engine.NewEngine(cfg, nil)
			_, err := eng.GenerateEvent(0)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: constraint validation failed: %v\n", err)
				os.Exit(2)
			}
		}

		fmt.Printf("✓ SUCCESS: '%s' is valid with %d fields and %d constraints.\n", *configPath, len(cfg.GetFields()), len(cfg.GetConstraints()))
		os.Exit(0)

	case "preview":
		previewCmd := flag.NewFlagSet("preview", flag.ExitOnError)
		configPath := previewCmd.String("config", "mimic.json", "Path to mimic.json")
		count := previewCmd.Int("count", 10, "Number of preview events")
		_ = previewCmd.Parse(os.Args[2:])

		cfg, err := loadConfig(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		eng := engine.NewEngine(cfg, nil)
		for i := 0; i < *count; i++ {
			evt, err := eng.GenerateEvent(int64(i))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Generation error at event %d: %v\n", i, err)
				os.Exit(1)
			}
			bytes, _ := json.MarshalIndent(evt, "", "  ")
			fmt.Println(string(bytes))
		}

	case "run":
		runCmd := flag.NewFlagSet("run", flag.ExitOnError)
		configPath := runCmd.String("config", "mimic.json", "Path to mimic.json")
		dryRun := runCmd.Bool("dry-run", false, "Simulate without sending to network sinks")
		_ = runCmd.Parse(os.Args[2:])

		cfg, err := loadConfig(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if *dryRun {
			cfg.Sink = &pb.SinkConfig{
				SinkType: &pb.SinkConfig_Stdout{
					Stdout: &pb.StdoutSink{Pretty: false},
				},
			}
			if cfg.EventLimit == 0 || cfg.EventLimit > 10 {
				cfg.EventLimit = 10
			}
			fmt.Printf("// DRY RUN: Producing %d events to stdout\n", cfg.EventLimit)
		}

		eng := engine.NewEngine(cfg, nil)
		if err := eng.Run(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
			os.Exit(1)
		}

	case "export":
		exportCmd := flag.NewFlagSet("export", flag.ExitOnError)
		templateName := exportCmd.String("to-template", "order-flow", "Template name (order-flow, user-signup)")
		_ = exportCmd.Parse(os.Args[2:])

		var tpl *pb.GeneratorConfig
		switch *templateName {
		case "order-flow":
			tpl = &pb.GeneratorConfig{
				Id:         "order-flow-template",
				Name:       "E-Commerce Order Flow",
				EventLimit: 1000,
				Sink: &pb.SinkConfig{
					SinkType: &pb.SinkConfig_Stdout{Stdout: &pb.StdoutSink{Pretty: true}},
				},
				Fields: []*pb.Field{
					{Name: "order_id", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Uuid{Uuid: &pb.UuidFunction{}}}},
					{Name: "customer_id", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_EntityPool{EntityPool: &pb.EntityPoolFunction{PoolSize: 100, EntityPrefix: "usr_"}}}},
					{Name: "amount", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Distribution{Distribution: &pb.DistributionFunction{DistributionType: &pb.DistributionFunction_Normal{Normal: &pb.NormalDistribution{Mean: 75.0, StdDev: 20.0}}}}}},
					{Name: "status", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_WeightedChoice{WeightedChoice: &pb.WeightedChoiceFunction{Options: []*pb.WeightedChoiceOption{{Value: "COMPLETED", Weight: 95.0}, {Value: "FAILED", Weight: 5.0}}}}}},
					{Name: "timestamp", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Now{Now: &pb.NowFunction{Format: "RFC3339"}}}},
				},
			}
		default:
			tpl = &pb.GeneratorConfig{
				Id:         "user-signup-template",
				Name:       "User Signups",
				EventLimit: 500,
				Sink:       &pb.SinkConfig{SinkType: &pb.SinkConfig_Stdout{Stdout: &pb.StdoutSink{Pretty: true}}},
				Fields: []*pb.Field{
					{Name: "user_id", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Uuid{Uuid: &pb.UuidFunction{}}}},
					{Name: "email", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_StringFaker{StringFaker: &pb.StringFakerFunction{Category: "email"}}}},
					{Name: "signup_at", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Now{Now: &pb.NowFunction{Format: "RFC3339"}}}},
				},
			}
		}

		bytes, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(tpl)
		fmt.Println(string(bytes))

	case "discover":
		discoverCmd := flag.NewFlagSet("discover", flag.ExitOnError)
		sqlFile := discoverCmd.String("sql", "", "Path to SQL file with CREATE TABLE")
		_ = discoverCmd.Parse(os.Args[2:])

		if *sqlFile == "" {
			fmt.Fprintln(os.Stderr, "Error: --sql argument is required")
			os.Exit(1)
		}

		sqlContent, err := os.ReadFile(*sqlFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading SQL file: %v\n", err)
			os.Exit(1)
		}

		cfg, err := importer.SQLDDLToGeneratorConfig(string(sqlContent))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing DDL: %v\n", err)
			os.Exit(1)
		}

		bytes, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(cfg)
		fmt.Println(string(bytes))

	case "cluster":
		clusterCmd := flag.NewFlagSet("cluster", flag.ExitOnError)
		workerCount := clusterCmd.Int("workers", 3, "Number of worker nodes to partition across")
		totalEvents := clusterCmd.Int64("total", 300000, "Total events goal to distribute")
		_ = clusterCmd.Parse(os.Args[2:])

		coord := cluster.NewCoordinator("cli-cluster")
		for i := 1; i <= *workerCount; i++ {
			_, _ = coord.Register(&pb.RegisterWorkerRequest{
				WorkerId: fmt.Sprintf("worker-%d", i),
				Hostname: fmt.Sprintf("node-%d.local", i),
				CpuCores: 8,
			})
		}

		plan, err := coord.PlanPartitions(*totalEvents)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cluster partitioning failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ SUCCESS: Distributed %d events across %d cluster workers:\n", *totalEvents, len(plan))
		for _, p := range plan {
			fmt.Printf("  • Worker %s: Event Range [%d - %d) -> %d events\n", p.WorkerID, p.StartIndex, p.EndIndex, p.Count)
		}

	case "profile":
		profileCmd := flag.NewFlagSet("profile", flag.ExitOnError)
		tableName := profileCmd.String("table", "profiled_data", "Table or entity name")
		sampleFile := profileCmd.String("sample", "", "Path to JSON file containing sample rows array")
		_ = profileCmd.Parse(os.Args[2:])

		if *sampleFile == "" {
			fmt.Fprintln(os.Stderr, "Error: --sample argument is required")
			os.Exit(1)
		}

		sampleBytes, err := os.ReadFile(*sampleFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading sample file: %v\n", err)
			os.Exit(1)
		}

		if len(sampleBytes) >= 3 && sampleBytes[0] == 0xEF && sampleBytes[1] == 0xBB && sampleBytes[2] == 0xBF {
			sampleBytes = sampleBytes[3:]
		}

		var rows []map[string]any
		if err := json.Unmarshal(sampleBytes, &rows); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing sample JSON array: %v\n", err)
			os.Exit(1)
		}

		p := profiler.NewDatabaseProfiler(*tableName, rows)
		cfg, err := p.Profile()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Profiling failed: %v\n", err)
			os.Exit(1)
		}

		bytes, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(cfg)
		fmt.Println(string(bytes))

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func loadConfig(path string) (*pb.GeneratorConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Strip UTF-8 BOM if present
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}

	var cfg pb.GeneratorConfig
	if err := protojson.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
