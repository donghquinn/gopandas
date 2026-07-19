package gopandas

import (
	"math"
	"sort"
)

// Describe returns a summary DataFrame with statistics for each numeric column.
// Rows: count, mean, std, min, 25%, 50%, 75%, max.
func (df *DataFrame) Describe() *DataFrame {
	numericCols := []string{}
	numericIndices := []int{}

	for i, col := range df.columns {
		for _, row := range df.data {
			if row[i] != nil {
				if _, ok := toFloat64(row[i]); ok {
					numericCols = append(numericCols, col)
					numericIndices = append(numericIndices, i)
				}
				break
			}
		}
	}

	allCols := append([]string{"stat"}, numericCols...)
	result := NewDataFrame(allCols)

	stats := []string{"count", "mean", "std", "min", "25%", "50%", "75%", "max"}

	// Collect and sort each numeric column once, then compute all stats from it.
	colStats := make([][]float64, len(numericIndices))
	for j, colIdx := range numericIndices {
		values := collectNumeric(df, colIdx)
		sort.Float64s(values)
		colStats[j] = computeStats(values)
	}

	result.data = make([][]interface{}, len(stats))
	result.index = make([]interface{}, len(stats))
	for i, stat := range stats {
		row := make([]interface{}, len(allCols))
		row[0] = stat
		for j := range numericIndices {
			row[j+1] = colStats[j][i]
		}
		result.data[i] = row
		result.index[i] = stat
	}

	return result
}

func collectNumeric(df *DataFrame, colIdx int) []float64 {
	values := make([]float64, 0, len(df.data))
	for _, row := range df.data {
		if f, ok := toFloat64(row[colIdx]); ok {
			values = append(values, f)
		}
	}
	return values
}

// computeStats returns count, mean, std, min, 25%, 50%, 75%, max for sorted values.
func computeStats(sorted []float64) []float64 {
	if len(sorted) == 0 {
		return make([]float64, 8)
	}

	m := mean64(sorted)

	var std float64
	if len(sorted) >= 2 {
		var ss float64
		for _, v := range sorted {
			d := v - m
			ss += d * d
		}
		std = math.Sqrt(ss / float64(len(sorted)-1))
	}

	return []float64{
		float64(len(sorted)),
		m,
		std,
		sorted[0],
		percentile64(sorted, 25),
		percentile64(sorted, 50),
		percentile64(sorted, 75),
		sorted[len(sorted)-1],
	}
}

func mean64(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func percentile64(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	idx := p / 100.0 * float64(n-1)
	lo := int(idx)
	hi := lo + 1
	if hi >= n {
		return sorted[lo]
	}
	frac := idx - float64(lo)
	return sorted[lo]*(1-frac) + sorted[hi]*frac
}
