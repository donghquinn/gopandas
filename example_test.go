package gopandas

import (
	"os"
	"testing"
)

func TestReadExcel(t *testing.T) {
	data, err := ReadExcel("excel.xlsx")
	if err != nil {
		t.Errorf("read excel err: %v", err)
	}
	if data == nil {
		t.Error("read data is nil")
	}
}

func TestReadCsv(t *testing.T) {
	data, err := ReadCSV("data.csv")
	if err != nil {
		t.Errorf("read csv err: %v", err)
	}
	if data == nil {
		t.Error("read data is nil")
	}
}

func TestDataFrameBasics(t *testing.T) {
	df := NewDataFrame([]string{"name", "age", "city"})

	rows, cols := df.Shape()
	if rows != 0 || cols != 3 {
		t.Errorf("Expected shape (0, 3), got (%d, %d)", rows, cols)
	}

	err := df.AddRow([]interface{}{"Alice", 25, "New York"})
	if err != nil {
		t.Errorf("Failed to add row: %v", err)
	}

	err = df.AddRow([]interface{}{"Bob", 30, "London"})
	if err != nil {
		t.Errorf("Failed to add row: %v", err)
	}

	rows, cols = df.Shape()
	if rows != 2 || cols != 3 {
		t.Errorf("Expected shape (2, 3), got (%d, %d)", rows, cols)
	}
}

func TestSeries(t *testing.T) {
	data := []interface{}{1, 2, 3, 4, 5}
	series := NewSeries("numbers", data)

	if series.Count() != 5 {
		t.Errorf("Expected count 5, got %d", series.Count())
	}

	sum, err := series.Sum()
	if err != nil {
		t.Errorf("Failed to calculate sum: %v", err)
	}

	if sum != 15.0 {
		t.Errorf("Expected sum 15, got %v", sum)
	}

	mean, err := series.Mean()
	if err != nil {
		t.Errorf("Failed to calculate mean: %v", err)
	}

	if mean != 3.0 {
		t.Errorf("Expected mean 3, got %v", mean)
	}
}

func TestCSVOperations(t *testing.T) {
	testData := "name,age,city\nAlice,25,New York\nBob,30,London\nCharlie,35,Paris\n"

	file, err := os.CreateTemp("", "test*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())

	_, err = file.WriteString(testData)
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	file.Close()

	df, err := ReadCSV(file.Name())
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	rows, cols := df.Shape()
	if rows != 3 || cols != 3 {
		t.Errorf("Expected shape (3, 3), got (%d, %d)", rows, cols)
	}

	columns := df.Columns()
	expected := []string{"name", "age", "city"}
	for i, col := range columns {
		if col != expected[i] {
			t.Errorf("Expected column %s, got %s", expected[i], col)
		}
	}
}

func TestDataFrameOperations(t *testing.T) {
	df := NewDataFrame([]string{"name", "age", "salary"})
	df.AddRow([]interface{}{"Alice", 25, 50000})
	df.AddRow([]interface{}{"Bob", 30, 60000})
	df.AddRow([]interface{}{"Charlie", 35, 70000})

	filtered := df.Filter(func(row []interface{}) bool {
		age := row[1].(int)
		return age >= 30
	})

	rows, cols := filtered.Shape()
	if rows != 2 || cols != 3 {
		t.Errorf("Expected filtered shape (2, 3), got (%d, %d)", rows, cols)
	}

	selected, err := df.Select("name", "age")
	if err != nil {
		t.Errorf("Failed to select columns: %v", err)
	}

	rows, cols = selected.Shape()
	if rows != 3 || cols != 2 {
		t.Errorf("Expected selected shape (3, 2), got (%d, %d)", rows, cols)
	}

	sorted, err := df.Sort("age", true)
	if err != nil {
		t.Errorf("Failed to sort: %v", err)
	}

	rows, cols = sorted.Shape()
	if rows != 3 || cols != 3 {
		t.Errorf("Expected sorted shape (3, 3), got (%d, %d)", rows, cols)
	}
}

func TestGroupBy(t *testing.T) {
	df := NewDataFrame([]string{"department", "salary"})
	df.AddRow([]interface{}{"Engineering", 70000})
	df.AddRow([]interface{}{"Sales", 50000})
	df.AddRow([]interface{}{"Engineering", 80000})
	df.AddRow([]interface{}{"Sales", 55000})

	groups, err := df.GroupBy("department")
	if err != nil {
		t.Errorf("Failed to group by: %v", err)
	}

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}

	engGroup := groups["Engineering"]
	rows, cols := engGroup.Shape()
	if rows != 2 || cols != 2 {
		t.Errorf("Expected Engineering group shape (2, 2), got (%d, %d)", rows, cols)
	}
}
