package fidelity

import (
	"fmt"
	"math"
	"sort"
)

// SummaryReport captures KS test, KL divergence approximation, and correlation fidelity.
type SummaryReport struct {
	FieldName    string
	KSDistance   float64
	KSPValuePass bool
	MeanDiff     float64
	StdDevDiff   float64
}

// KolmogorovSmirnovTest compares candidate synthetic samples against a reference dataset.
// Returns maximum vertical distance D between empirical CDFs.
func KolmogorovSmirnovTest(reference, synthetic []float64) float64 {
	if len(reference) == 0 || len(synthetic) == 0 {
		return 1.0
	}

	refSorted := make([]float64, len(reference))
	copy(refSorted, reference)
	sort.Float64s(refSorted)

	synSorted := make([]float64, len(synthetic))
	copy(synSorted, synthetic)
	sort.Float64s(synSorted)

	n1 := float64(len(refSorted))
	n2 := float64(len(synSorted))

	var maxD float64
	i, j := 0, 0

	for i < len(refSorted) && j < len(synSorted) {
		val1 := refSorted[i]
		val2 := synSorted[j]

		var currentVal float64
		if val1 <= val2 {
			currentVal = val1
			i++
		} else {
			currentVal = val2
			j++
		}

		// Calculate empirical CDF at currentVal
		cdf1 := float64(i) / n1
		cdf2 := float64(j) / n2
		d := math.Abs(cdf1 - cdf2)
		if d > maxD {
			maxD = d
		}
		_ = currentVal
	}

	return maxD
}

// PearsonCorrelation computes linear correlation coefficient between two numeric fields.
func PearsonCorrelation(x, y []float64) (float64, error) {
	if len(x) != len(y) || len(x) == 0 {
		return 0, fmt.Errorf("sample arrays must have equal positive lengths")
	}

	n := float64(len(x))
	var sumX, sumY, sumX2, sumY2, sumXY float64

	for i := 0; i < len(x); i++ {
		sumX += x[i]
	}
	for i := 0; i < len(y); i++ {
		sumY += y[i]
	}

	meanX := sumX / n
	meanY := sumY / n

	for i := 0; i < len(x); i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		sumXY += dx * dy
		sumX2 += dx * dx
		sumY2 += dy * dy
	}

	denom := math.Sqrt(sumX2 * sumY2)
	if denom == 0 {
		return 0, nil
	}

	return sumXY / denom, nil
}

