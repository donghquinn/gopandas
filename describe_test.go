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
