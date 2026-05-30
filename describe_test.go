package gopandas

import (
	"testing"
)

func TestDescribe(t *testing.T) {
	df := NewDataFrame([]string{"val"})
	for _, v := range []interface{}{1, 2, 3, 4, 5} {
		df.AddRow([]interface{}{v})
	}

	desc := df.Describe()

	// Shape: 8 stat rows, 2 cols (stat + val)
	rows, cols := desc.Shape()
	if rows != 8 {
		t.Errorf("Describe: expected 8 rows, got %d", rows)
	}
	if cols != 2 {
		t.Errorf("Describe: expected 2 cols, got %d", cols)
	}

	// Retrieve val column values (one float64 per stat row)
	valCol, err := desc.GetColumn("val")
	if err != nil {
		t.Fatalf("Describe: GetColumn(val): %v", err)
	}
	vals := valCol.Values()

	// Row 0 = count = 5
	if vals[0].(float64) != 5.0 {
		t.Errorf("Describe count: expected 5.0, got %v", vals[0])
	}
	// Row 1 = mean = 3
	if vals[1].(float64) != 3.0 {
		t.Errorf("Describe mean: expected 3.0, got %v", vals[1])
	}
	// Row 3 = min = 1
	if vals[3].(float64) != 1.0 {
		t.Errorf("Describe min: expected 1.0, got %v", vals[3])
	}
	// Row 7 = max = 5
	if vals[7].(float64) != 5.0 {
		t.Errorf("Describe max: expected 5.0, got %v", vals[7])
	}
	// Row 4 = 25% = 2
	if vals[4].(float64) != 2.0 {
		t.Errorf("Describe 25%%: expected 2.0, got %v", vals[4])
	}
	// Row 5 = 50% = 3
	if vals[5].(float64) != 3.0 {
		t.Errorf("Describe 50%%: expected 3.0, got %v", vals[5])
	}
	// Row 6 = 75% = 4
	if vals[6].(float64) != 4.0 {
		t.Errorf("Describe 75%%: expected 4.0, got %v", vals[6])
	}
}

func TestDescribeMultipleNumericColumns(t *testing.T) {
	df := NewDataFrame([]string{"a", "b"})
	df.AddRow([]interface{}{1, 10})
	df.AddRow([]interface{}{2, 20})
	df.AddRow([]interface{}{3, 30})

	desc := df.Describe()

	// Shape: 8 stat rows, 3 cols (stat + a + b)
	rows, cols := desc.Shape()
	if rows != 8 {
		t.Errorf("Describe multi: expected 8 rows, got %d", rows)
	}
	if cols != 3 {
		t.Errorf("Describe multi: expected 3 cols, got %d", cols)
	}
}

func TestDescribeSkipsStringColumns(t *testing.T) {
	df := NewDataFrame([]string{"name", "age"})
	df.AddRow([]interface{}{"Alice", 25})
	df.AddRow([]interface{}{"Bob", 30})

	desc := df.Describe()

	// "name" is a string column; it should not appear in describe result
	cols := desc.Columns()
	for _, col := range cols {
		if col == "name" {
			t.Error("Describe: string column 'name' should not be in result")
		}
	}

	// "age" should be present
	found := false
	for _, col := range cols {
		if col == "age" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Describe: numeric column 'age' should be in result")
	}
}

// TestDescribeWithNullsInNumericColumn covers the nil-skip (continue) in collectNumeric.
func TestDescribeWithNullsInNumericColumn(t *testing.T) {
	df := NewDataFrame([]string{"val"})
	df.AddRow([]interface{}{10})
	df.AddRow([]interface{}{nil}) // nil in numeric column → collectNumeric skips it
	df.AddRow([]interface{}{20})
	df.AddRow([]interface{}{30})

	desc := df.Describe()
	r, _ := desc.Shape()
	if r != 8 {
		t.Errorf("Describe with nulls: expected 8 rows, got %d", r)
	}
	// count should be 3 (nil excluded)
	valCol, _ := desc.GetColumn("val")
	if valCol.Values()[0].(float64) != 3.0 {
		t.Errorf("Describe with nulls count: expected 3.0, got %v", valCol.Values()[0])
	}
}

// TestDescribeFloat64Values covers the float64 arm in collectNumeric.
func TestDescribeFloat64Values(t *testing.T) {
	df := NewDataFrame([]string{"val"})
	for _, v := range []interface{}{1.0, 2.0, 3.0, 4.0, 5.0} {
		df.AddRow([]interface{}{v})
	}
	desc := df.Describe()
	r, _ := desc.Shape()
	if r != 8 {
		t.Errorf("Describe float64: expected 8 stat rows, got %d", r)
	}
	valCol, _ := desc.GetColumn("val")
	if valCol.Values()[0].(float64) != 5.0 { // count
		t.Errorf("Describe float64 count: expected 5.0, got %v", valCol.Values()[0])
	}
}

// TestDescribeFloat32Values covers the float32 arm in collectNumeric.
func TestDescribeFloat32Values(t *testing.T) {
	df := NewDataFrame([]string{"val"})
	for _, v := range []interface{}{float32(2), float32(4), float32(6)} {
		df.AddRow([]interface{}{v})
	}
	desc := df.Describe()
	r, _ := desc.Shape()
	if r != 8 {
		t.Errorf("Describe float32: expected 8 stat rows, got %d", r)
	}
	valCol, _ := desc.GetColumn("val")
	if valCol.Values()[0].(float64) != 3.0 { // count=3
		t.Errorf("Describe float32 count: expected 3.0, got %v", valCol.Values()[0])
	}
}

// TestDescribeSingleRow covers the percentile64 hi>=n edge case (single element).
func TestDescribeSingleRow(t *testing.T) {
	df := NewDataFrame([]string{"val"})
	df.AddRow([]interface{}{42})

	desc := df.Describe()
	r, _ := desc.Shape()
	if r != 8 {
		t.Errorf("Describe single row: expected 8 stat rows, got %d", r)
	}
	valCol, _ := desc.GetColumn("val")
	// std of a single value must be 0 (len < 2 guard)
	std := valCol.Values()[2].(float64)
	if std != 0.0 {
		t.Errorf("Describe single row std: expected 0.0, got %v", std)
	}
	// min and max both equal 42
	min := valCol.Values()[3].(float64)
	max := valCol.Values()[7].(float64)
	if min != 42.0 || max != 42.0 {
		t.Errorf("Describe single row min/max: expected 42/42, got %v/%v", min, max)
	}
}

func TestDescribeStatRowOrder(t *testing.T) {
	df := NewDataFrame([]string{"val"})
	for _, v := range []interface{}{1, 2, 3, 4, 5} {
		df.AddRow([]interface{}{v})
	}

	desc := df.Describe()

	statCol, err := desc.GetColumn("stat")
	if err != nil {
		t.Fatalf("Describe stat row order: GetColumn(stat): %v", err)
	}
	vals := statCol.Values()

	expectedOrder := []string{"count", "mean", "std", "min", "25%", "50%", "75%", "max"}
	if len(vals) != len(expectedOrder) {
		t.Fatalf("Describe stat rows: expected %d, got %d", len(expectedOrder), len(vals))
	}
	for i, expected := range expectedOrder {
		got, ok := vals[i].(string)
		if !ok {
			t.Errorf("stat[%d]: expected string, got %T", i, vals[i])
			continue
		}
		if got != expected {
			t.Errorf("stat[%d]: expected %q, got %q", i, expected, got)
		}
	}
}
