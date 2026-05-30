package gopandas

import (
	"math"
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

func TestTailAndIloc(t *testing.T) {
	df := NewDataFrame([]string{"n"})
	for i := 1; i <= 5; i++ {
		df.AddRow([]interface{}{i})
	}

	tail := df.Tail(2)
	r, _ := tail.Shape()
	if r != 2 {
		t.Errorf("Tail: expected 2 rows, got %d", r)
	}

	slice, err := df.Iloc(1, 3)
	if err != nil {
		t.Fatalf("Iloc error: %v", err)
	}
	r, _ = slice.Shape()
	if r != 2 {
		t.Errorf("Iloc: expected 2 rows, got %d", r)
	}

	rows, err := df.IlocRows(0, 4)
	if err != nil {
		t.Fatalf("IlocRows error: %v", err)
	}
	r, _ = rows.Shape()
	if r != 2 {
		t.Errorf("IlocRows: expected 2 rows, got %d", r)
	}
}

func TestRenameAndDrop(t *testing.T) {
	df := NewDataFrame([]string{"a", "b", "c"})
	df.AddRow([]interface{}{1, 2, 3})

	renamed := df.Rename(map[string]string{"a": "x", "c": "z"})
	cols := renamed.Columns()
	if cols[0] != "x" || cols[1] != "b" || cols[2] != "z" {
		t.Errorf("Rename failed: %v", cols)
	}

	dropped, err := df.Drop("b")
	if err != nil {
		t.Fatalf("Drop error: %v", err)
	}
	_, c := dropped.Shape()
	if c != 2 {
		t.Errorf("Drop: expected 2 cols, got %d", c)
	}
}

func TestNullHandling(t *testing.T) {
	df := NewDataFrame([]string{"a", "b"})
	df.AddRow([]interface{}{1, nil})
	df.AddRow([]interface{}{2, 3})
	df.AddRow([]interface{}{nil, 4})

	clean := df.DropNA()
	r, _ := clean.Shape()
	if r != 1 {
		t.Errorf("DropNA: expected 1 row, got %d", r)
	}

	filled := df.FillNA(0)
	for _, row := range filled.data {
		for _, val := range row {
			if val == nil {
				t.Error("FillNA: found nil after fill")
			}
		}
	}

	s := NewSeries("x", []interface{}{1, nil, 3, nil})
	nulls := s.IsNull()
	if !nulls[1] || !nulls[3] || nulls[0] {
		t.Error("IsNull failed")
	}
	clean2 := s.DropNA()
	if clean2.Count() != 2 {
		t.Errorf("Series DropNA: expected 2, got %d", clean2.Count())
	}
	filled2 := s.FillNA(0)
	if filled2.Count() != 4 {
		t.Errorf("Series FillNA: expected 4, got %d", filled2.Count())
	}
}

func TestConcat(t *testing.T) {
	df1 := NewDataFrame([]string{"a", "b"})
	df1.AddRow([]interface{}{1, 2})
	df2 := NewDataFrame([]string{"a", "b"})
	df2.AddRow([]interface{}{3, 4})
	df2.AddRow([]interface{}{5, 6})

	combined, err := Concat(df1, df2)
	if err != nil {
		t.Fatalf("Concat error: %v", err)
	}
	r, c := combined.Shape()
	if r != 3 || c != 2 {
		t.Errorf("Concat: expected (3,2), got (%d,%d)", r, c)
	}
}

func TestMerge(t *testing.T) {
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
		t.Fatalf("Merge inner error: %v", err)
	}
	r, _ := inner.Shape()
	if r != 2 {
		t.Errorf("Merge inner: expected 2 rows, got %d", r)
	}

	leftJoin, err := left.Merge(right, "id", "left")
	if err != nil {
		t.Fatalf("Merge left error: %v", err)
	}
	r, _ = leftJoin.Shape()
	if r != 3 {
		t.Errorf("Merge left: expected 3 rows, got %d", r)
	}

	outer, err := left.Merge(right, "id", "outer")
	if err != nil {
		t.Fatalf("Merge outer error: %v", err)
	}
	r, _ = outer.Shape()
	if r != 4 {
		t.Errorf("Merge outer: expected 4 rows, got %d", r)
	}
}

func TestSeriesStats(t *testing.T) {
	s := NewSeries("nums", []interface{}{2, 4, 4, 4, 5, 5, 7, 9})

	max, err := s.Max()
	if err != nil || max.(int) != 9 {
		t.Errorf("Max: expected 9, got %v (err %v)", max, err)
	}
	min, err := s.Min()
	if err != nil || min.(int) != 2 {
		t.Errorf("Min: expected 2, got %v (err %v)", min, err)
	}
	median, err := s.Median()
	if err != nil || median != 4.5 {
		t.Errorf("Median: expected 4.5, got %v (err %v)", median, err)
	}
	std, err := s.Std()
	if err != nil {
		t.Fatalf("Std error: %v", err)
	}
	// sample std of {2,4,4,4,5,5,7,9} = sqrt(32/7) ≈ 2.138
	if math.Abs(std-2.138) > 0.01 {
		t.Errorf("Std: expected ~2.138, got %v", std)
	}

	unique := s.Unique()
	if len(unique) != 5 {
		t.Errorf("Unique: expected 5, got %d", len(unique))
	}
	if s.NUnique() != 5 {
		t.Errorf("NUnique: expected 5, got %d", s.NUnique())
	}

	vc := s.ValueCounts()
	if vc[4] != 3 {
		t.Errorf("ValueCounts: expected 4→3, got %d", vc[4])
	}

	doubled := s.Apply(func(v interface{}) interface{} {
		return v.(int) * 2
	})
	if doubled.Values()[0].(int) != 4 {
		t.Errorf("Apply: expected 4, got %v", doubled.Values()[0])
	}
}

func TestDescribe(t *testing.T) {
	df := NewDataFrame([]string{"val"})
	for _, v := range []interface{}{1, 2, 3, 4, 5} {
		df.AddRow([]interface{}{v})
	}
	desc := df.Describe()
	r, _ := desc.Shape()
	if r != 8 {
		t.Errorf("Describe: expected 8 rows, got %d", r)
	}
}

func TestJSONRoundtrip(t *testing.T) {
	df := NewDataFrame([]string{"name", "age"})
	df.AddRow([]interface{}{"Alice", 25})
	df.AddRow([]interface{}{"Bob", 30})

	f, err := os.CreateTemp("", "test*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	if err := df.ToJSON(f.Name()); err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}

	df2, err := ReadJSON(f.Name())
	if err != nil {
		t.Fatalf("ReadJSON error: %v", err)
	}
	r, c := df2.Shape()
	if r != 2 || c != 2 {
		t.Errorf("ReadJSON: expected (2,2), got (%d,%d)", r, c)
	}
}
