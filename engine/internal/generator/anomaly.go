package generator

import (
	"strings"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

// AnomalyInjector applies scheduled anomalies (corrupt payload, null fields, timeouts) across candidate records.
type AnomalyInjector struct {
	playbooks []*pb.AnomalyPlaybook
}

func NewAnomalyInjector(playbooks []*pb.AnomalyPlaybook) *AnomalyInjector {
	return &AnomalyInjector{playbooks: playbooks}
}

// Apply determines whether an anomaly should be injected for the current event.
func (a *AnomalyInjector) Apply(record map[string]any, rng *state.DeterministicRand) (map[string]any, bool) {
	if len(a.playbooks) == 0 {
		return record, false
	}

	mutated := false
	for _, pbItem := range a.playbooks {
		rate := pbItem.GetAnomalyRate()
		if rate <= 0 {
			rate = 0.05 // default 5%
		}

		// Probability check via deterministic RNG
		if rng.Float64() < rate {
			mutated = true
			switch strings.ToLower(pbItem.GetInjectionType()) {
			case "null_field":
				// Pick a key and set to nil
				for k := range record {
					record[k] = nil
					break
				}
			case "corrupt_payload":
				record["_corrupt"] = true
				record["_raw_malformed_bytes"] = "0xFF0xFE0x00_MALFORMED"
			case "burst":
				record["_burst_spike"] = true
			default:
				// Fallback: add anomaly flag
				record["_anomaly"] = pbItem.GetName()
			}
		}
	}

	return record, mutated
}

