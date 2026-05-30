package gopandas

import (
	"testing"
)

func TestConcat(t *testing.T) {
	// Two frames combined rows
	df1 := NewDataFrame([]string{"a", "b"})
	df1.AddRow([]interface{}{1, 2})
	df2 := NewDataFrame([]string{"a", "b"})
	df2.AddRow([]interface{}{3, 4})
	df2.AddRow([]interface{}{5, 6})

	combined, err := Concat(df1, df2)
	if err != nil {
		t.Fatalf("Concat two frames: unexpected error: %v", err)
	}
	rows, cols := combined.Shape()
	if rows != 3 || cols != 2 {
		t.Errorf("Concat two frames: expected (3,2), got (%d,%d)", rows, cols)
	}

	// Three frames
	df3 := NewDataFrame([]string{"a", "b"})
	df3.AddRow([]interface{}{7, 8})

	three, err := Concat(df1, df2, df3)
	if err != nil {
		t.Fatalf("Concat three frames: unexpected error: %v", err)
	}
	rows, _ = three.Shape()
	if rows != 4 {
		t.Errorf("Concat three frames: expected 4 rows, got %d", rows)
	}
}

func TestConcatEmpty(t *testing.T) {
	_, err := Concat()
	if err == nil {
		t.Error("Concat no args: expected error, got nil")
	}
}

func TestConcatMismatchedColumns(t *testing.T) {
	df1 := NewDataFrame([]string{"a", "b"})
	df1.AddRow([]interface{}{1, 2})
	df2 := NewDataFrame([]string{"a", "c"}) // different column name
	df2.AddRow([]interface{}{3, 4})

	_, err := Concat(df1, df2)
	if err == nil {
		t.Error("Concat mismatched column names: expected error, got nil")
	}
}

func TestConcatColumnCountMismatch(t *testing.T) {
	df1 := NewDataFrame([]string{"a", "b"})
	df1.AddRow([]interface{}{1, 2})
	df2 := NewDataFrame([]string{"a", "b", "c"}) // different column count
	df2.AddRow([]interface{}{3, 4, 5})

	_, err := Concat(df1, df2)
	if err == nil {
		t.Error("Concat different column count: expected error, got nil")
	}
}

func TestMergeInner(t *testing.T) {
	left := NewDataFrame([]string{"id", "name"})
	left.AddRow([]interface{}{1, "Alice"})
	left.AddRow([]interface{}{2, "Bob"})
	left.AddRow([]interface{}{3, "Charlie"})

	right := NewDataFrame([]string{"id", "score"})
	right.AddRow([]interface{}{1, 90})
	right.AddRow([]interface{}{2, 85})
	right.AddRow([]interface{}{4, 70})

	inner, err := left.Merge(right, "id", "inner")
	if err != nil {
		t.Fatalf("Merge inner: unexpected error: %v", err)
	}
	rows, cols := inner.Shape()
	if rows != 2 {
		t.Errorf("Merge inner: expected 2 rows, got %d", rows)
	}
	if cols != 3 {
		t.Errorf("Merge inner: expected 3 cols (id, name, score), got %d", cols)
	}
}

func TestMergeLeft(t *testing.T) {
	left := NewDataFrame([]string{"id", "name"})
	left.AddRow([]interface{}{1, "Alice"})
	left.AddRow([]interface{}{2, "Bob"})
	left.AddRow([]interface{}{3, "Charlie"})

	right := NewDataFrame([]string{"id", "score"})
	right.AddRow([]interface{}{1, 90})
	right.AddRow([]interface{}{4, 70})

	result, err := left.Merge(right, "id", "left")
	if err != nil {
		t.Fatalf("Merge left: unexpected error: %v", err)
	}
	rows, _ := result.Shape()
	if rows != 3 {
		t.Errorf("Merge left: expected 3 rows (all left), got %d", rows)
	}

	// Unmatched right values should be nil
	scoreCol, err := result.GetColumn("score")
	if err != nil {
		t.Fatalf("Merge left: GetColumn score: %v", err)
	}
	scoreVals := scoreCol.Values()
	if scoreVals[1] != nil {
		t.Errorf("Merge left: score[1] expected nil for unmatched, got %v", scoreVals[1])
	}
}

func TestMergeRight(t *testing.T) {
	left := NewDataFrame([]string{"id", "name"})
	left.AddRow([]interface{}{1, "Alice"})
	left.AddRow([]interface{}{2, "Bob"})

	right := NewDataFrame([]string{"id", "score"})
	right.AddRow([]interface{}{1, 90})
	right.AddRow([]interface{}{2, 85})
	right.AddRow([]interface{}{3, 70})

	result, err := left.Merge(right, "id", "right")
	if err != nil {
		t.Fatalf("Merge right: unexpected error: %v", err)
	}
	rows, _ := result.Shape()
	if rows != 3 {
		t.Errorf("Merge right: expected 3 rows (all right), got %d", rows)
	}
}

func TestMergeOuter(t *testing.T) {
	left := NewDataFrame([]string{"id", "name"})
	left.AddRow([]interface{}{1, "Alice"})
	left.AddRow([]interface{}{2, "Bob"})

	right := NewDataFrame([]string{"id", "score"})
	right.AddRow([]interface{}{2, 85})
	right.AddRow([]interface{}{3, 70})

	result, err := left.Merge(right, "id", "outer")
	if err != nil {
		t.Fatalf("Merge outer: unexpected error: %v", err)
	}
	rows, _ := result.Shape()
	// ids: 1 (left only), 2 (both), 3 (right only) → 3 rows
	if rows != 3 {
		t.Errorf("Merge outer: expected 3 rows (union), got %d", rows)
	}
}

func TestMergeInvalidHow(t *testing.T) {
	left := NewDataFrame([]string{"id", "val"})
	left.AddRow([]interface{}{1, "a"})
	right := NewDataFrame([]string{"id", "val"})
	right.AddRow([]interface{}{1, "b"})

	_, err := left.Merge(right, "id", "invalid")
	if err == nil {
		t.Error("Merge invalid how: expected error, got nil")
	}
}

func TestMergeMissingKey(t *testing.T) {
	left := NewDataFrame([]string{"id", "name"})
	left.AddRow([]interface{}{1, "Alice"})
	right := NewDataFrame([]string{"id", "score"})
	right.AddRow([]interface{}{1, 90})

	// Key not in left DataFrame
	_, err := left.Merge(right, "nonexistent", "inner")
	if err == nil {
		t.Error("Merge missing key: expected error, got nil")
	}
}

func TestMergeColumnConflict(t *testing.T) {
	// Both DataFrames have same non-key column name
	left := NewDataFrame([]string{"id", "val"})
	left.AddRow([]interface{}{1, "left_val"})
	left.AddRow([]interface{}{2, "left_val2"})

	right := NewDataFrame([]string{"id", "val"})
	right.AddRow([]interface{}{1, "right_val"})
	right.AddRow([]interface{}{2, "right_val2"})

	result, err := left.Merge(right, "id", "inner")
	if err != nil {
		t.Fatalf("Merge column conflict: unexpected error: %v", err)
	}

	cols := result.Columns()
	// Expected: [id, val, val_right]
	if len(cols) != 3 {
		t.Fatalf("Merge column conflict: expected 3 cols, got %d: %v", len(cols), cols)
	}
	if cols[0] != "id" {
		t.Errorf("col[0] expected id, got %q", cols[0])
	}
	if cols[1] != "val" {
		t.Errorf("col[1] expected val, got %q", cols[1])
	}
	if cols[2] != "val_right" {
		t.Errorf("col[2] expected val_right, got %q", cols[2])
	}
}
