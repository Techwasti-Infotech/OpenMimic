package telemetry

import (
	"strings"
	"testing"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

func TestPrometheusScrapeOutput(t *testing.T) {
	exp := GetDefaultExporter()
	exp.Record("gen-test", "kafka", 1500, 2, 4.25, 120.0)

	metricsText := exp.Scrape()

	if !strings.Contains(metricsText, "openmimic_events_generated_total{generator_id=\"gen-test\",sink=\"kafka\"} 1500") {
		t.Fatalf("expected events count in metrics, got:\n%s", metricsText)
	}
	if !strings.Contains(metricsText, "openmimic_throughput_eps{generator_id=\"gen-test\",sink=\"kafka\"} 120.00") {
		t.Fatalf("expected throughput in metrics, got:\n%s", metricsText)
	}
}

func TestOTelTraceparentInjection(t *testing.T) {
	cfg := &pb.OTelTracingConfig{
		Enabled:     true,
		ServiceName: "ecom-producer",
		SampleRate:  1.0,
	}

	inj := NewOTelInjector(cfg)
	rng := state.NewDeterministicRand(42)

	record := map[string]any{"event": "login"}
	injected := inj.Inject(record, rng)

	traceparent, ok := injected["_traceparent"].(string)
	if !ok || !strings.HasPrefix(traceparent, "00-") {
		t.Fatalf("expected valid W3C traceparent, got %v", injected["_traceparent"])
	}

	if injected["_otel_service"] != "ecom-producer" {
		t.Fatalf("expected service name ecom-producer, got %v", injected["_otel_service"])
	}
}

