package gopandas

import (
	"math"
	"testing"
)

func TestSeriesMax(t *testing.T) {
	// Int series
	si := NewSeries("ints", []interface{}{3, 1, 4, 1, 5, 9, 2, 6})
	max, err := si.Max()
	if err != nil {
		t.Fatalf("SeriesMax ints: unexpected error: %v", err)
	}
	if max.(int) != 9 {
		t.Errorf("SeriesMax ints: expected 9, got %v", max)
	}

	// Float64 series
	sf := NewSeries("floats", []interface{}{1.1, 3.3, 2.2})
	max, err = sf.Max()
	if err != nil {
		t.Fatalf("SeriesMax floats: unexpected error: %v", err)
	}
	if max.(float64) != 3.3 {
		t.Errorf("SeriesMax floats: expected 3.3, got %v", max)
	}

	// String series (lexicographic)
	ss := NewSeries("strings", []interface{}{"banana", "apple", "cherry"})
	max, err = ss.Max()
	if err != nil {
		t.Fatalf("SeriesMax strings: unexpected error: %v", err)
	}
	if max.(string) != "cherry" {
		t.Errorf("SeriesMax strings: expected cherry, got %v", max)
	}
}

func TestSeriesMin(t *testing.T) {
	// Int series
	si := NewSeries("ints", []interface{}{3, 1, 4, 1, 5, 9, 2, 6})
	min, err := si.Min()
	if err != nil {
		t.Fatalf("SeriesMin ints: unexpected error: %v", err)
	}
	if min.(int) != 1 {
		t.Errorf("SeriesMin ints: expected 1, got %v", min)
	}

	// Float64 series
	sf := NewSeries("floats", []interface{}{1.1, 3.3, 2.2})
	min, err = sf.Min()
	if err != nil {
		t.Fatalf("SeriesMin floats: unexpected error: %v", err)
	}
	if min.(float64) != 1.1 {
		t.Errorf("SeriesMin floats: expected 1.1, got %v", min)
	}

	// String series
	ss := NewSeries("strings", []interface{}{"banana", "apple", "cherry"})
	min, err = ss.Min()
	if err != nil {
		t.Fatalf("SeriesMin strings: unexpected error: %v", err)
	}
	if min.(string) != "apple" {
		t.Errorf("SeriesMin strings: expected apple, got %v", min)
	}
}

func TestSeriesMaxWithNils(t *testing.T) {
	s := NewSeries("nums", []interface{}{nil, 3, nil, 7, nil, 2})
	max, err := s.Max()
	if err != nil {
		t.Fatalf("SeriesMaxWithNils: unexpected error: %v", err)
	}
	if max.(int) != 7 {
		t.Errorf("SeriesMaxWithNils: expected 7, got %v", max)
	}
}

func TestSeriesMaxEmpty(t *testing.T) {
	s := NewSeries("empty", []interface{}{})
	_, err := s.Max()
	if err == nil {
		t.Error("SeriesMaxEmpty: expected error for empty series, got nil")
	}
}

func TestSeriesVar(t *testing.T) {
	// Sample variance of {2,4,4,4,5,5,7,9} = 32/7 ≈ 4.571
	s := NewSeries("nums", []interface{}{2, 4, 4, 4, 5, 5, 7, 9})
	v, err := s.Var()
	if err != nil {
		t.Fatalf("SeriesVar: unexpected error: %v", err)
	}
	expected := 32.0 / 7.0
	if math.Abs(v-expected) > 0.001 {
		t.Errorf("SeriesVar: expected ~%.4f, got %.4f", expected, v)
	}
}

func TestSeriesVarInsufficientData(t *testing.T) {
	s := NewSeries("single", []interface{}{42})
	_, err := s.Var()
	if err == nil {
		t.Error("SeriesVarInsufficientData: expected error for single element, got nil")
	}
}

func TestSeriesStd(t *testing.T) {
	// std of {2,4,4,4,5,5,7,9} = sqrt(32/7) ≈ 2.138
	s := NewSeries("nums", []interface{}{2, 4, 4, 4, 5, 5, 7, 9})
	std, err := s.Std()
	if err != nil {
		t.Fatalf("SeriesStd: unexpected error: %v", err)
	}
	expected := math.Sqrt(32.0 / 7.0)
	if math.Abs(std-expected) > 0.01 {
		t.Errorf("SeriesStd: expected ~%.4f, got %.4f", expected, std)
	}
}

func TestSeriesMedianOdd(t *testing.T) {
	s := NewSeries("nums", []interface{}{1, 3, 5})
	median, err := s.Median()
	if err != nil {
		t.Fatalf("SeriesMedianOdd: unexpected error: %v", err)
	}
	if median != 3.0 {
		t.Errorf("SeriesMedianOdd: expected 3.0, got %v", median)
	}
}

func TestSeriesMedianEven(t *testing.T) {
	s := NewSeries("nums", []interface{}{1, 2, 3, 4})
	median, err := s.Median()
	if err != nil {
		t.Fatalf("SeriesMedianEven: unexpected error: %v", err)
	}
	if median != 2.5 {
		t.Errorf("SeriesMedianEven: expected 2.5, got %v", median)
	}
}

func TestSeriesMedianEmpty(t *testing.T) {
	s := NewSeries("empty", []interface{}{})
	_, err := s.Median()
	if err == nil {
		t.Error("SeriesMedianEmpty: expected error for empty series, got nil")
	}
}

func TestSeriesUnique(t *testing.T) {
	s := NewSeries("nums", []interface{}{1, 2, 1, 3, 2, 4})
	unique := s.Unique()

	if len(unique) != 4 {
		t.Errorf("SeriesUnique: expected 4 unique values, got %d", len(unique))
	}
	// Order preserved (first-seen)
	if unique[0].(int) != 1 {
		t.Errorf("SeriesUnique: unique[0] expected 1, got %v", unique[0])
	}
	if unique[1].(int) != 2 {
		t.Errorf("SeriesUnique: unique[1] expected 2, got %v", unique[1])
	}
	if unique[2].(int) != 3 {
		t.Errorf("SeriesUnique: unique[2] expected 3, got %v", unique[2])
	}
	if unique[3].(int) != 4 {
		t.Errorf("SeriesUnique: unique[3] expected 4, got %v", unique[3])
	}
}

func TestSeriesNUnique(t *testing.T) {
	// Count of distinct non-nil values; nil excluded
	s := NewSeries("nums", []interface{}{1, 2, 1, nil, 3, nil, 2})
	n := s.NUnique()
	if n != 3 {
		t.Errorf("SeriesNUnique: expected 3, got %d", n)
	}
}

func TestSeriesValueCounts(t *testing.T) {
	s := NewSeries("nums", []interface{}{1, 2, 1, 2, 2, nil, 3})
	vc := s.ValueCounts()

	// nil excluded
	if _, hasNil := vc[nil]; hasNil {
		t.Error("SeriesValueCounts: nil should be excluded from map")
	}
	if vc[1] != 2 {
		t.Errorf("SeriesValueCounts: 1 expected count 2, got %d", vc[1])
	}
	if vc[2] != 3 {
		t.Errorf("SeriesValueCounts: 2 expected count 3, got %d", vc[2])
	}
	if vc[3] != 1 {
		t.Errorf("SeriesValueCounts: 3 expected count 1, got %d", vc[3])
	}
}

func TestSeriesApplyStats(t *testing.T) {
	s := NewSeries("nums", []interface{}{1, 2, 3, 4, 5})
	doubled := s.Apply(func(v interface{}) interface{} {
		return v.(int) * 2
	})

	vals := doubled.Values()
	if len(vals) != 5 {
		t.Fatalf("SeriesApplyStats: expected 5 values, got %d", len(vals))
	}
	if vals[0].(int) != 2 {
		t.Errorf("SeriesApplyStats: vals[0] expected 2, got %v", vals[0])
	}
	if vals[4].(int) != 10 {
		t.Errorf("SeriesApplyStats: vals[4] expected 10, got %v", vals[4])
	}
}

func TestSeriesName(t *testing.T) {
	s := NewSeries("my_series", []interface{}{1, 2, 3})
	if s.Name() != "my_series" {
		t.Errorf("SeriesName: expected my_series, got %q", s.Name())
	}
}

func TestSeriesValues(t *testing.T) {
	data := []interface{}{10, 20, 30}
	s := NewSeries("test", data)
	vals := s.Values()

	if len(vals) != 3 {
		t.Fatalf("SeriesValues: expected 3 values, got %d", len(vals))
	}
	for i, v := range data {
		if vals[i] != v {
			t.Errorf("SeriesValues[%d]: expected %v, got %v", i, v, vals[i])
		}
	}
}

// TestSeriesMinEmpty covers the error return in Min (all-nil series).
func TestSeriesMinEmpty(t *testing.T) {
	s := NewSeries("empty", []interface{}{nil, nil})
	_, err := s.Min()
	if err == nil {
		t.Error("Min of all-nil series: expected error, got nil")
	}
}

// TestSeriesStdSingleElement covers the Std error path when Var fails (< 2 values).
func TestSeriesStdSingleElement(t *testing.T) {
	s := NewSeries("one", []interface{}{42})
	_, err := s.Std()
	if err == nil {
		t.Error("Std of single-element series: expected error, got nil")
	}
}

// TestSeriesMedianWithNils covers the nil-skip (continue) in numericFloats.
func TestSeriesMedianWithNils(t *testing.T) {
	s := NewSeries("mixed", []interface{}{1, nil, 3, nil, 5})
	median, err := s.Median()
	if err != nil {
		t.Fatalf("MedianWithNils: %v", err)
	}
	if median != 3.0 {
		t.Errorf("MedianWithNils: expected 3.0, got %v", median)
	}
}

// TestSeriesVarWithNils ensures nil values are skipped in Var (also exercises numericFloats nil path).
func TestSeriesVarWithNils(t *testing.T) {
	s := NewSeries("mixed", []interface{}{2, nil, 4, nil, 6})
	v, err := s.Var()
	if err != nil {
		t.Fatalf("VarWithNils: %v", err)
	}
	if v == 0 {
		t.Error("VarWithNils: expected non-zero variance")
	}
}

// TestSeriesFloat64Ops covers the float64 arm in numericFloats.
func TestSeriesFloat64Ops(t *testing.T) {
	s := NewSeries("f64", []interface{}{1.5, 2.5, 3.5, 4.5})
	sum, err := s.Sum()
	if err != nil {
		t.Fatalf("float64 Sum: %v", err)
	}
	if sum.(float64) != 12.0 {
		t.Errorf("float64 Sum: expected 12.0, got %v", sum)
	}
	mean, err := s.Mean()
	if err != nil {
		t.Fatalf("float64 Mean: %v", err)
	}
	if mean != 3.0 {
		t.Errorf("float64 Mean: expected 3.0, got %v", mean)
	}
	median, err := s.Median()
	if err != nil {
		t.Fatalf("float64 Median: %v", err)
	}
	if median != 3.0 {
		t.Errorf("float64 Median: expected 3.0, got %v", median)
	}
}

// TestSeriesFloat32Ops covers the float32 arm in numericFloats.
func TestSeriesFloat32Ops(t *testing.T) {
	s := NewSeries("f32", []interface{}{float32(2), float32(4), float32(6)})
	sum, err := s.Sum()
	if err != nil {
		t.Fatalf("float32 Sum: %v", err)
	}
	if sum.(float64) != 12.0 {
		t.Errorf("float32 Sum: expected 12.0, got %v", sum)
	}
	mean, err := s.Mean()
	if err != nil {
		t.Fatalf("float32 Mean: %v", err)
	}
	if mean != 4.0 {
		t.Errorf("float32 Mean: expected 4.0, got %v", mean)
	}
	std, err := s.Std()
	if err != nil {
		t.Fatalf("float32 Std: %v", err)
	}
	if std == 0 {
		t.Error("float32 Std: expected non-zero")
	}
}
