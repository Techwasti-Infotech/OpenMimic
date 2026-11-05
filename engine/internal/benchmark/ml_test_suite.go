package benchmark

import (
	"fmt"
	"math"
	"sort"
)

// AnomalyScoreResult summarizes precision, recall, and detection accuracy of injected anomalies.
type AnomalyScoreResult struct {
	TotalSamples     int     `json:"total_samples"`
	GroundTruthAnoms int     `json:"ground_truth_anomalies"`
	DetectedAnoms    int     `json:"detected_anomalies"`
	TruePositives    int     `json:"true_positives"`
	FalsePositives   int     `json:"false_positives"`
	Precision        float64 `json:"precision"`
	Recall           float64 `json:"recall"`
	F1Score          float64 `json:"f1_score"`
}

// IsolationScoreBenchmark computes contamination scores using an empirical quantile isolation estimator.
func IsolationScoreBenchmark(values []float64, groundTruthInjected []bool, contamination float64) (*AnomalyScoreResult, error) {
	n := len(values)
	if n == 0 || len(groundTruthInjected) != n {
		return nil, fmt.Errorf("invalid inputs: values and truth slices must have identical non-zero length")
	}

	// Compute sample mean and standard deviation
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(n)

	varianceSum := 0.0
	for _, v := range values {
		diff := v - mean
		varianceSum += diff * diff
	}
	stdDev := math.Sqrt(varianceSum / float64(n))
	if stdDev == 0 {
		stdDev = 1.0
	}

	// Calculate anomaly Z-score |v - mean| / stdDev
	scores := make([]float64, n)
	sortedScores := make([]float64, n)
	for i, v := range values {
		score := math.Abs(v-mean) / stdDev
		scores[i] = score
		sortedScores[i] = score
	}
	sort.Float64s(sortedScores)

	// Threshold at (1 - contamination) quantile
	cutoffIdx := int(float64(n) * (1.0 - contamination))
	if cutoffIdx >= n {
		cutoffIdx = n - 1
	}
	threshold := sortedScores[cutoffIdx]

	tp, fp, fn := 0, 0, 0
	groundTruthCount := 0
	detectedCount := 0

	for i := 0; i < n; i++ {
		isRealAnomaly := groundTruthInjected[i]
		if isRealAnomaly {
			groundTruthCount++
		}
		isDetected := scores[i] >= threshold
		if isDetected {
			detectedCount++
			if isRealAnomaly {
				tp++
			} else {
				fp++
			}
		} else if isRealAnomaly {
			fn++
		}
	}

	precision := 0.0
	if tp+fp > 0 {
		precision = float64(tp) / float64(tp+fp)
	}
	recall := 0.0
	if tp+fn > 0 {
		recall = float64(tp) / float64(tp+fn)
	}
	f1 := 0.0
	if precision+recall > 0 {
		f1 = 2.0 * (precision * recall) / (precision + recall)
	}

	return &AnomalyScoreResult{
		TotalSamples:     n,
		GroundTruthAnoms: groundTruthCount,
		DetectedAnoms:    detectedCount,
		TruePositives:    tp,
		FalsePositives:   fp,
		Precision:        precision,
		Recall:           recall,
		F1Score:          f1,
	}, nil
}

