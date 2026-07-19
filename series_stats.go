package gopandas

import (
	"fmt"
	"math"
	"sort"
)

// Max returns the maximum value in the Series.
func (s *Series) Max() (interface{}, error) {
	var maxVal interface{}
	for _, val := range s.data {
		if val == nil {
			continue
		}
		if maxVal == nil || compareValues(val, maxVal) > 0 {
			maxVal = val
		}
	}
	if maxVal == nil {
		return nil, fmt.Errorf("no non-null values found")
	}
	return maxVal, nil
}

// Min returns the minimum value in the Series.
func (s *Series) Min() (interface{}, error) {
	var minVal interface{}
	for _, val := range s.data {
		if val == nil {
			continue
		}
		if minVal == nil || compareValues(val, minVal) < 0 {
			minVal = val
		}
	}
	if minVal == nil {
		return nil, fmt.Errorf("no non-null values found")
	}
	return minVal, nil
}

// Var returns the sample variance (ddof=1) of numeric values.
func (s *Series) Var() (float64, error) {
	values := s.numericFloats()
	if len(values) < 2 {
		return 0, fmt.Errorf("need at least 2 numeric values to compute variance")
	}
	m := mean64(values)
	var ss float64
	for _, v := range values {
		d := v - m
		ss += d * d
	}
	return ss / float64(len(values)-1), nil
}

// Std returns the sample standard deviation (ddof=1) of numeric values.
func (s *Series) Std() (float64, error) {
	v, err := s.Var()
	if err != nil {
		return 0, err
	}
	return math.Sqrt(v), nil
}

// Median returns the median of numeric values.
func (s *Series) Median() (float64, error) {
	values := s.numericFloats()
	if len(values) == 0 {
		return 0, fmt.Errorf("no numeric values found")
	}
	sort.Float64s(values)
	n := len(values)
	if n%2 == 0 {
		return (values[n/2-1] + values[n/2]) / 2, nil
	}
	return values[n/2], nil
}

// Unique returns the distinct values in the Series, preserving first-seen order.
func (s *Series) Unique() []interface{} {
	seen := make(map[interface{}]bool, len(s.data))
	result := make([]interface{}, 0, len(s.data))
	for _, val := range s.data {
		if !seen[val] {
			seen[val] = true
			result = append(result, val)
		}
	}
	return result
}

// NUnique returns the number of distinct non-nil values.
func (s *Series) NUnique() int {
	seen := make(map[interface{}]bool)
	for _, val := range s.data {
		if val != nil {
			seen[val] = true
		}
	}
	return len(seen)
}

// ValueCounts returns a map of value → occurrence count, excluding nil.
func (s *Series) ValueCounts() map[interface{}]int {
	counts := make(map[interface{}]int)
	for _, val := range s.data {
		if val != nil {
			counts[val]++
		}
	}
	return counts
}

// Apply applies fn to each element and returns a new Series.
func (s *Series) Apply(fn func(interface{}) interface{}) *Series {
	newData := make([]interface{}, len(s.data))
	for i, val := range s.data {
		newData[i] = fn(val)
	}
	return NewSeries(s.name, newData)
}

// Name returns the name of the Series.
func (s *Series) Name() string {
	return s.name
}

// Values returns the raw data slice.
func (s *Series) Values() []interface{} {
	return s.data
}

func (s *Series) numericFloats() []float64 {
	values := make([]float64, 0, len(s.data))
	for _, val := range s.data {
		if f, ok := toFloat64(val); ok {
			values = append(values, f)
		}
	}
	return values
}
