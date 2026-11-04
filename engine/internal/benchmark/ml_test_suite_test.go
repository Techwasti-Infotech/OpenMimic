package benchmark

import (
	"math/rand"
	"testing"
)

func TestAnomalyMLDetectionBenchmark(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	n := 1000
	values := make([]float64, n)
	groundTruth := make([]bool, n)

	// Generate baseline normal distribution N(100, 10)
	for i := 0; i < n; i++ {
		values[i] = 100.0 + rng.NormFloat64()*10.0
		groundTruth[i] = false
	}

	// Inject 30 synthetic extreme outliers (OpenMimic anomaly injection simulation)
	anomIndices := []int{25, 80, 140, 205, 310, 450, 520, 680, 790, 895, 950}
	for _, idx := range anomIndices {
		values[idx] = 250.0 + rng.Float64()*50.0 // Extreme spike
		groundTruth[idx] = true
	}

	result, err := IsolationScoreBenchmark(values, groundTruth, 0.02)
	if err != nil {
		t.Fatalf("unexpected benchmark error: %v", err)
	}

	if result.TruePositives == 0 {
		t.Errorf("expected ML benchmark to detect true positive anomalies")
	}

	if result.Recall < 0.70 {
		t.Errorf("expected recall >= 0.70, got %f", result.Recall)
	}

	if result.F1Score < 0.60 {
		t.Errorf("expected F1-score >= 0.60, got %f", result.F1Score)
	}
}
