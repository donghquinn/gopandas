package gopandas

import (
	"testing"
)

func TestDataFrameDropNA(t *testing.T) {
	// Frame with no nulls unchanged
	df := NewDataFrame([]string{"a", "b"})
	df.AddRow([]interface{}{1, 2})
	df.AddRow([]interface{}{3, 4})

	clean := df.DropNA()
	rows, _ := clean.Shape()
	if rows != 2 {
		t.Errorf("DropNA no nulls: expected 2 rows, got %d", rows)
	}

	// Frame with partial nulls removes only those rows
	df2 := NewDataFrame([]string{"a", "b"})
	df2.AddRow([]interface{}{1, nil})
	df2.AddRow([]interface{}{2, 3})
	df2.AddRow([]interface{}{nil, 4})

	clean2 := df2.DropNA()
	rows, _ = clean2.Shape()
	if rows != 1 {
		t.Errorf("DropNA partial nulls: expected 1 row, got %d", rows)
	}

	// All-null rows removed
	df3 := NewDataFrame([]string{"x", "y"})
	df3.AddRow([]interface{}{nil, nil})
	df3.AddRow([]interface{}{nil, nil})

	clean3 := df3.DropNA()
	rows, _ = clean3.Shape()
	if rows != 0 {
		t.Errorf("DropNA all null rows: expected 0 rows, got %d", rows)
	}
}

func TestDataFrameFillNA(t *testing.T) {
	df := NewDataFrame([]string{"a", "b"})
	df.AddRow([]interface{}{1, nil})
	df.AddRow([]interface{}{nil, 3})

	filled := df.FillNA(0)

	// Check shape unchanged
	rows, cols := filled.Shape()
	if rows != 2 || cols != 2 {
		t.Fatalf("FillNA: shape changed, got (%d,%d)", rows, cols)
	}

	// Nils replaced by fill value
	colA, _ := filled.GetColumn("a")
	colB, _ := filled.GetColumn("b")
	avs := colA.Values()
	bvs := colB.Values()

	if avs[1] == nil {
		t.Error("FillNA: a[1] should not be nil")
	}
	if avs[1].(int) != 0 {
		t.Errorf("FillNA: a[1] expected 0, got %v", avs[1])
	}
	if bvs[0] == nil {
		t.Error("FillNA: b[0] should not be nil")
	}
	if bvs[0].(int) != 0 {
		t.Errorf("FillNA: b[0] expected 0, got %v", bvs[0])
	}

	// Non-nil values untouched
	if avs[0].(int) != 1 {
		t.Errorf("FillNA: a[0] should remain 1, got %v", avs[0])
	}
	if bvs[1].(int) != 3 {
		t.Errorf("FillNA: b[1] should remain 3, got %v", bvs[1])
	}
}

func TestDataFrameHasNA(t *testing.T) {
	// Returns true when nulls present
	df := NewDataFrame([]string{"a", "b"})
	df.AddRow([]interface{}{1, nil})
	if !df.HasNA() {
		t.Error("HasNA: expected true for frame with nil, got false")
	}

	// Returns false when clean
	clean := NewDataFrame([]string{"a", "b"})
	clean.AddRow([]interface{}{1, 2})
	if clean.HasNA() {
		t.Error("HasNA: expected false for clean frame, got true")
	}
}

func TestSeriesIsNull(t *testing.T) {
	s := NewSeries("x", []interface{}{1, nil, 3, nil, 5})
	nulls := s.IsNull()

	if len(nulls) != 5 {
		t.Fatalf("IsNull: expected length 5, got %d", len(nulls))
	}
	if nulls[0] {
		t.Error("IsNull[0]: expected false, got true")
	}
	if !nulls[1] {
		t.Error("IsNull[1]: expected true, got false")
	}
	if nulls[2] {
		t.Error("IsNull[2]: expected false, got true")
	}
	if !nulls[3] {
		t.Error("IsNull[3]: expected true, got false")
	}
	if nulls[4] {
		t.Error("IsNull[4]: expected false, got true")
	}
}

func TestSeriesNotNull(t *testing.T) {
	s := NewSeries("x", []interface{}{1, nil, 3, nil, 5})
	notNulls := s.NotNull()

	if len(notNulls) != 5 {
		t.Fatalf("NotNull: expected length 5, got %d", len(notNulls))
	}
	// Should be inverse of IsNull
	if !notNulls[0] {
		t.Error("NotNull[0]: expected true, got false")
	}
	if notNulls[1] {
		t.Error("NotNull[1]: expected false, got true")
	}
	if !notNulls[2] {
		t.Error("NotNull[2]: expected true, got false")
	}
	if notNulls[3] {
		t.Error("NotNull[3]: expected false, got true")
	}
	if !notNulls[4] {
		t.Error("NotNull[4]: expected true, got false")
	}
}

func TestSeriesDropNA(t *testing.T) {
	s := NewSeries("x", []interface{}{1, nil, 3, nil, 5})
	dropped := s.DropNA()

	vals := dropped.Values()
	if len(vals) != 3 {
		t.Fatalf("Series DropNA: expected 3 values, got %d", len(vals))
	}
	// Preserves order
	if vals[0].(int) != 1 {
		t.Errorf("Series DropNA: vals[0] expected 1, got %v", vals[0])
	}
	if vals[1].(int) != 3 {
		t.Errorf("Series DropNA: vals[1] expected 3, got %v", vals[1])
	}
	if vals[2].(int) != 5 {
		t.Errorf("Series DropNA: vals[2] expected 5, got %v", vals[2])
	}
}

func TestSeriesFillNA(t *testing.T) {
	s := NewSeries("x", []interface{}{1, nil, 3, nil, 5})
	filled := s.FillNA(0)

	vals := filled.Values()
	if len(vals) != 5 {
		t.Fatalf("Series FillNA: count should stay same (5), got %d", len(vals))
	}
	if vals[1].(int) != 0 {
		t.Errorf("Series FillNA: vals[1] expected 0, got %v", vals[1])
	}
	if vals[3].(int) != 0 {
		t.Errorf("Series FillNA: vals[3] expected 0, got %v", vals[3])
	}
	// Non-nil values preserved
	if vals[0].(int) != 1 {
		t.Errorf("Series FillNA: vals[0] expected 1, got %v", vals[0])
	}
	if vals[2].(int) != 3 {
		t.Errorf("Series FillNA: vals[2] expected 3, got %v", vals[2])
	}
}
