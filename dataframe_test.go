package gopandas

import (
	"strings"
	"testing"
)

// makePersonDF creates a DataFrame with columns: name, age, city; 3 rows.
func makePersonDF() *DataFrame {
	df := NewDataFrame([]string{"name", "age", "city"})
	df.AddRow([]interface{}{"Alice", 25, "New York"})
	df.AddRow([]interface{}{"Bob", 30, "London"})
	df.AddRow([]interface{}{"Charlie", 35, "Paris"})
	return df
}

// makeNumericDF creates a DataFrame with columns: a, b; 3 rows.
func makeNumericDF() *DataFrame {
	df := NewDataFrame([]string{"a", "b"})
	df.AddRow([]interface{}{1, 10})
	df.AddRow([]interface{}{2, 20})
	df.AddRow([]interface{}{3, 30})
	return df
}

func TestNewDataFrame(t *testing.T) {
	df := NewDataFrame([]string{"x", "y", "z"})
	rows, cols := df.Shape()
	if rows != 0 {
		t.Errorf("expected 0 rows, got %d", rows)
	}
	if cols != 3 {
		t.Errorf("expected 3 cols, got %d", cols)
	}
}

func TestAddRow(t *testing.T) {
	df := NewDataFrame([]string{"a", "b"})

	// Success case
	err := df.AddRow([]interface{}{1, 2})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	rows, _ := df.Shape()
	if rows != 1 {
		t.Errorf("expected 1 row after AddRow, got %d", rows)
	}

	// Error on wrong length
	err = df.AddRow([]interface{}{1})
	if err == nil {
		t.Error("expected error for wrong row length, got nil")
	}
}

func TestShape(t *testing.T) {
	df := NewDataFrame([]string{"a", "b", "c"})
	rows, cols := df.Shape()
	if rows != 0 || cols != 3 {
		t.Errorf("empty df: expected (0,3), got (%d,%d)", rows, cols)
	}

	df.AddRow([]interface{}{1, 2, 3})
	df.AddRow([]interface{}{4, 5, 6})
	df.AddRow([]interface{}{7, 8, 9})
	rows, cols = df.Shape()
	if rows != 3 || cols != 3 {
		t.Errorf("after 3 rows: expected (3,3), got (%d,%d)", rows, cols)
	}
}

func TestColumns(t *testing.T) {
	expected := []string{"x", "y", "z"}
	df := NewDataFrame(expected)
	cols := df.Columns()
	if len(cols) != len(expected) {
		t.Fatalf("expected %d columns, got %d", len(expected), len(cols))
	}
	for i, c := range cols {
		if c != expected[i] {
			t.Errorf("col[%d]: expected %q, got %q", i, expected[i], c)
		}
	}
}

func TestHead(t *testing.T) {
	df := makePersonDF()

	// n < len
	h := df.Head(2)
	rows, _ := h.Shape()
	if rows != 2 {
		t.Errorf("Head(2): expected 2 rows, got %d", rows)
	}

	// n > len (clamps to len)
	h = df.Head(10)
	rows, _ = h.Shape()
	if rows != 3 {
		t.Errorf("Head(10) on 3-row df: expected 3 rows, got %d", rows)
	}

	// n = 0
	h = df.Head(0)
	rows, _ = h.Shape()
	if rows != 0 {
		t.Errorf("Head(0): expected 0 rows, got %d", rows)
	}
}

func TestTail(t *testing.T) {
	df := makePersonDF()

	// n < len
	tail := df.Tail(2)
	rows, _ := tail.Shape()
	if rows != 2 {
		t.Errorf("Tail(2): expected 2 rows, got %d", rows)
	}

	// n > len
	tail = df.Tail(10)
	rows, _ = tail.Shape()
	if rows != 3 {
		t.Errorf("Tail(10) on 3-row df: expected 3 rows, got %d", rows)
	}
}

func TestGetColumn(t *testing.T) {
	df := makePersonDF()

	// Valid column
	s, err := df.GetColumn("name")
	if err != nil {
		t.Fatalf("GetColumn(name): unexpected error: %v", err)
	}
	vals := s.Values()
	if len(vals) != 3 {
		t.Fatalf("GetColumn(name): expected 3 values, got %d", len(vals))
	}
	if vals[0].(string) != "Alice" {
		t.Errorf("GetColumn(name)[0]: expected Alice, got %v", vals[0])
	}
	if vals[1].(string) != "Bob" {
		t.Errorf("GetColumn(name)[1]: expected Bob, got %v", vals[1])
	}

	// Invalid column
	_, err = df.GetColumn("nonexistent")
	if err == nil {
		t.Error("GetColumn(nonexistent): expected error, got nil")
	}
}

func TestRename(t *testing.T) {
	df := NewDataFrame([]string{"a", "b", "c"})
	df.AddRow([]interface{}{1, 2, 3})

	// Partial rename
	renamed := df.Rename(map[string]string{"a": "x", "c": "z"})
	cols := renamed.Columns()
	if cols[0] != "x" {
		t.Errorf("expected col[0]=x, got %q", cols[0])
	}
	if cols[1] != "b" {
		t.Errorf("expected col[1]=b (unchanged), got %q", cols[1])
	}
	if cols[2] != "z" {
		t.Errorf("expected col[2]=z, got %q", cols[2])
	}
}

func TestDropColumn(t *testing.T) {
	df := makePersonDF()

	// Single column
	dropped, err := df.Drop("city")
	if err != nil {
		t.Fatalf("Drop(city): unexpected error: %v", err)
	}
	_, cols := dropped.Shape()
	if cols != 2 {
		t.Errorf("Drop(city): expected 2 cols, got %d", cols)
	}

	// Multiple columns
	dropped2, err := df.Drop("age", "city")
	if err != nil {
		t.Fatalf("Drop(age,city): unexpected error: %v", err)
	}
	_, cols = dropped2.Shape()
	if cols != 1 {
		t.Errorf("Drop(age,city): expected 1 col, got %d", cols)
	}

	// Nonexistent column
	_, err = df.Drop("nonexistent")
	if err == nil {
		t.Error("Drop(nonexistent): expected error, got nil")
	}
}

func TestIloc(t *testing.T) {
	df := makePersonDF()

	// [1,3) returns 2 rows
	result, err := df.Iloc(1, 3)
	if err != nil {
		t.Fatalf("Iloc(1,3): unexpected error: %v", err)
	}
	rows, _ := result.Shape()
	if rows != 2 {
		t.Errorf("Iloc(1,3): expected 2 rows, got %d", rows)
	}

	// Out of range returns error
	_, err = df.Iloc(0, 10)
	if err == nil {
		t.Error("Iloc(0,10): expected error, got nil")
	}

	// Start > end returns error
	_, err = df.Iloc(3, 1)
	if err == nil {
		t.Error("Iloc(3,1): expected error for start>end, got nil")
	}
}

func TestIlocRows(t *testing.T) {
	df := makePersonDF()

	// Specific indices
	result, err := df.IlocRows(0, 2)
	if err != nil {
		t.Fatalf("IlocRows(0,2): unexpected error: %v", err)
	}
	rows, _ := result.Shape()
	if rows != 2 {
		t.Errorf("IlocRows(0,2): expected 2 rows, got %d", rows)
	}

	// Out of range returns error
	_, err = df.IlocRows(5)
	if err == nil {
		t.Error("IlocRows(5): expected error for out-of-range index, got nil")
	}
}

func TestDataFrameApply(t *testing.T) {
	df := makeNumericDF()

	applied := df.Apply(func(row []interface{}) []interface{} {
		newRow := make([]interface{}, len(row))
		for i, v := range row {
			newRow[i] = v.(int) * 2
		}
		return newRow
	})

	rows, _ := applied.Shape()
	if rows != 3 {
		t.Fatalf("Apply: expected 3 rows, got %d", rows)
	}

	colA, err := applied.GetColumn("a")
	if err != nil {
		t.Fatalf("Apply GetColumn: %v", err)
	}
	vals := colA.Values()
	if vals[0].(int) != 2 {
		t.Errorf("Apply: expected a[0]=2, got %v", vals[0])
	}
	if vals[1].(int) != 4 {
		t.Errorf("Apply: expected a[1]=4, got %v", vals[1])
	}
}

func TestDtypes(t *testing.T) {
	df := NewDataFrame([]string{"ints", "strs", "nils"})
	df.AddRow([]interface{}{1, "hello", nil})
	df.AddRow([]interface{}{2, "world", nil})

	dtypes := df.Dtypes()
	if dtypes["ints"] != "int" {
		t.Errorf("Dtypes: ints col expected int, got %q", dtypes["ints"])
	}
	if dtypes["strs"] != "string" {
		t.Errorf("Dtypes: strs col expected string, got %q", dtypes["strs"])
	}
	if dtypes["nils"] != "unknown" {
		t.Errorf("Dtypes: nils col expected unknown, got %q", dtypes["nils"])
	}
}

func TestIndex(t *testing.T) {
	df := makePersonDF()
	idx := df.Index()
	if len(idx) != 3 {
		t.Fatalf("Index: expected 3 entries, got %d", len(idx))
	}
	for i, v := range idx {
		if v.(int) != i {
			t.Errorf("Index[%d]: expected %d, got %v", i, i, v)
		}
	}
}

func TestSetColumn(t *testing.T) {
	df := makeNumericDF()

	// Replace existing column
	err := df.SetColumn("a", []interface{}{10, 20, 30})
	if err != nil {
		t.Fatalf("SetColumn replace: unexpected error: %v", err)
	}
	colA, _ := df.GetColumn("a")
	vals := colA.Values()
	if vals[0].(int) != 10 {
		t.Errorf("SetColumn replace: expected a[0]=10, got %v", vals[0])
	}

	// Add new column
	err = df.SetColumn("c", []interface{}{100, 200, 300})
	if err != nil {
		t.Fatalf("SetColumn add: unexpected error: %v", err)
	}
	_, cols := df.Shape()
	if cols != 3 {
		t.Errorf("SetColumn add: expected 3 cols, got %d", cols)
	}

	// Length mismatch returns error
	err = df.SetColumn("a", []interface{}{1, 2})
	if err == nil {
		t.Error("SetColumn length mismatch: expected error, got nil")
	}
}

// TestIlocNegativeEnd covers the negative end-index arm in Iloc.
func TestIlocNegativeEnd(t *testing.T) {
	df := makePersonDF() // 3 rows: Alice, Bob, Charlie

	// Iloc(-3, -1) → start=0, end=2 → rows [0,2) = Alice, Bob
	result, err := df.Iloc(-3, -1)
	if err != nil {
		t.Fatalf("Iloc negative end: %v", err)
	}
	r, _ := result.Shape()
	if r != 2 {
		t.Errorf("Iloc negative end: expected 2 rows, got %d", r)
	}
	nameCol, _ := result.GetColumn("name")
	if nameCol.Values()[0] != "Alice" {
		t.Errorf("Iloc negative end: expected Alice at row 0, got %v", nameCol.Values()[0])
	}
}

func TestString(t *testing.T) {
	df := makePersonDF()
	s := df.String()
	if s == "" {
		t.Error("String: expected non-empty output")
	}
	for _, col := range df.Columns() {
		if !strings.Contains(s, col) {
			t.Errorf("String: expected output to contain column name %q", col)
		}
	}
}
