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
				switch row[i].(type) {
				case int, float64, float32:
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

	for _, stat := range stats {
		row := make([]interface{}, len(allCols))
		row[0] = stat

		for j, colIdx := range numericIndices {
			values := collectNumeric(df, colIdx)
			row[j+1] = computeStat(stat, values)
		}

		result.data = append(result.data, row)
		result.index = append(result.index, stat)
	}

	return result
}

func collectNumeric(df *DataFrame, colIdx int) []float64 {
	values := make([]float64, 0, len(df.data))
	for _, row := range df.data {
		v := row[colIdx]
		if v == nil {
			continue
		}
		switch n := v.(type) {
		case int:
			values = append(values, float64(n))
		case float64:
			values = append(values, n)
		case float32:
			values = append(values, float64(n))
		}
	}
	return values
}

func computeStat(stat string, values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	switch stat {
	case "count":
		return float64(len(values))
	case "mean":
		return mean64(values)
	case "std":
		if len(values) < 2 {
			return 0
		}
		m := mean64(values)
		var ss float64
		for _, v := range values {
			d := v - m
			ss += d * d
		}
		return math.Sqrt(ss / float64(len(values)-1))
	case "min":
		s := sorted64(values)
		return s[0]
	case "25%":
		return percentile64(sorted64(values), 25)
	case "50%":
		return percentile64(sorted64(values), 50)
	case "75%":
		return percentile64(sorted64(values), 75)
	case "max":
		s := sorted64(values)
		return s[len(s)-1]
	}
	return 0
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

func sorted64(values []float64) []float64 {
	cp := make([]float64, len(values))
	copy(cp, values)
	sort.Float64s(cp)
	return cp
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
