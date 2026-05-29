package pairs

import "math"

func Returns(values []float64) []float64 {
	out := make([]float64, len(values))
	for i := 1; i < len(values); i++ {
		if values[i-1] == 0 {
			continue
		}
		out[i] = (values[i] - values[i-1]) / values[i-1]
	}
	return out
}

func Pearson(a []float64, b []float64) float64 {
	n := min(len(a), len(b))
	if n == 0 {
		return 0
	}

	meanA := 0.0
	meanB := 0.0
	for i := 0; i < n; i++ {
		meanA += a[i]
		meanB += b[i]
	}
	meanA /= float64(n)
	meanB /= float64(n)

	num := 0.0
	denA := 0.0
	denB := 0.0
	for i := 0; i < n; i++ {
		da := a[i] - meanA
		db := b[i] - meanB
		num += da * db
		denA += da * da
		denB += db * db
	}

	den := math.Sqrt(denA * denB)
	if den == 0 {
		return 0
	}
	return num / den
}
