package gopandas

import (
	"testing"
)

func TestFilter(t *testing.T) {
	df := makePersonDF() // name, age, city; Alice/25, Bob/30, Charlie/35

	// age >= 30 → 2 rows
	filtered := df.Filter(func(row []interface{}) bool {
		return row[1].(int) >= 30
	})
	rows, _ := filtered.Shape()
	if rows != 2 {
		t.Errorf("Filter age>=30: expected 2 rows, got %d", rows)
	}

	// impossible predicate → 0 rows
	none := df.Filter(func(row []interface{}) bool {
		return false
	})
	rows, _ = none.Shape()
	if rows != 0 {
		t.Errorf("Filter impossible: expected 0 rows, got %d", rows)
	}

	// always-true → all rows
	all := df.Filter(func(row []interface{}) bool {
		return true
	})
	rows, _ = all.Shape()
	if rows != 3 {
		t.Errorf("Filter all: expected 3 rows, got %d", rows)
	}
}

func TestSelect(t *testing.T) {
	df := makePersonDF()

	// Subset of columns
	sel, err := df.Select("name", "age")
	if err != nil {
		t.Fatalf("Select(name,age): unexpected error: %v", err)
	}
	rows, cols := sel.Shape()
	if rows != 3 || cols != 2 {
		t.Errorf("Select(name,age): expected (3,2), got (%d,%d)", rows, cols)
	}

	// Single column
	sel, err = df.Select("city")
	if err != nil {
		t.Fatalf("Select(city): unexpected error: %v", err)
	}
	_, cols = sel.Shape()
	if cols != 1 {
		t.Errorf("Select(city): expected 1 col, got %d", cols)
	}

	// Nonexistent column returns error
	_, err = df.Select("nonexistent")
	if err == nil {
		t.Error("Select(nonexistent): expected error, got nil")
	}
}

func TestSortAscending(t *testing.T) {
	df := makeNumericDF() // a: 1,2,3

	sorted, err := df.Sort("a", true)
	if err != nil {
		t.Fatalf("Sort ascending: unexpected error: %v", err)
	}
	colA, _ := sorted.GetColumn("a")
	vals := colA.Values()
	if vals[0].(int) != 1 {
		t.Errorf("Sort asc: expected first=1, got %v", vals[0])
	}
	if vals[len(vals)-1].(int) != 3 {
		t.Errorf("Sort asc: expected last=3, got %v", vals[len(vals)-1])
	}
}

func TestSortDescending(t *testing.T) {
	df := makeNumericDF() // a: 1,2,3

	sorted, err := df.Sort("a", false)
	if err != nil {
		t.Fatalf("Sort descending: unexpected error: %v", err)
	}
	colA, _ := sorted.GetColumn("a")
	vals := colA.Values()
	if vals[0].(int) != 3 {
		t.Errorf("Sort desc: expected first=3, got %v", vals[0])
	}
	if vals[len(vals)-1].(int) != 1 {
		t.Errorf("Sort desc: expected last=1, got %v", vals[len(vals)-1])
	}
}

func TestSortStrings(t *testing.T) {
	df := NewDataFrame([]string{"word"})
	df.AddRow([]interface{}{"banana"})
	df.AddRow([]interface{}{"apple"})
	df.AddRow([]interface{}{"cherry"})

	// Ascending
	sorted, err := df.Sort("word", true)
	if err != nil {
		t.Fatalf("SortStrings asc: unexpected error: %v", err)
	}
	col, _ := sorted.GetColumn("word")
	vals := col.Values()
	if vals[0].(string) != "apple" {
		t.Errorf("SortStrings asc: expected first=apple, got %v", vals[0])
	}
	if vals[len(vals)-1].(string) != "cherry" {
		t.Errorf("SortStrings asc: expected last=cherry, got %v", vals[len(vals)-1])
	}

	// Descending
	sorted, err = df.Sort("word", false)
	if err != nil {
		t.Fatalf("SortStrings desc: unexpected error: %v", err)
	}
	col, _ = sorted.GetColumn("word")
	vals = col.Values()
	if vals[0].(string) != "cherry" {
		t.Errorf("SortStrings desc: expected first=cherry, got %v", vals[0])
	}
}

func TestSortInvalidColumn(t *testing.T) {
	df := makePersonDF()
	_, err := df.Sort("nonexistent", true)
	if err == nil {
		t.Error("Sort invalid column: expected error, got nil")
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
		t.Fatalf("GroupBy: unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Errorf("GroupBy: expected 2 groups, got %d", len(groups))
	}

	eng := groups["Engineering"]
	rows, _ := eng.Shape()
	if rows != 2 {
		t.Errorf("GroupBy Engineering: expected 2 rows, got %d", rows)
	}

	sales := groups["Sales"]
	rows, _ = sales.Shape()
	if rows != 2 {
		t.Errorf("GroupBy Sales: expected 2 rows, got %d", rows)
	}
}

func TestGroupByInvalidColumn(t *testing.T) {
	df := makePersonDF()
	_, err := df.GroupBy("nonexistent")
	if err == nil {
		t.Error("GroupBy invalid column: expected error, got nil")
	}
}

func TestSeriesSum(t *testing.T) {
	// Int values
	s := NewSeries("ints", []interface{}{1, 2, 3, 4, 5})
	sum, err := s.Sum()
	if err != nil {
		t.Fatalf("Sum ints: unexpected error: %v", err)
	}
	if sum.(float64) != 15.0 {
		t.Errorf("Sum ints: expected 15.0, got %v", sum)
	}

	// Float64 values
	sf := NewSeries("floats", []interface{}{1.5, 2.5, 3.0})
	sum, err = sf.Sum()
	if err != nil {
		t.Fatalf("Sum floats: unexpected error: %v", err)
	}
	if sum.(float64) != 7.0 {
		t.Errorf("Sum floats: expected 7.0, got %v", sum)
	}

	// With nil values (nils skipped)
	sn := NewSeries("with_nils", []interface{}{1, nil, 2, nil, 3})
	sum, err = sn.Sum()
	if err != nil {
		t.Fatalf("Sum with nils: unexpected error: %v", err)
	}
	if sum.(float64) != 6.0 {
		t.Errorf("Sum with nils: expected 6.0, got %v", sum)
	}

	// All nil returns error
	sall := NewSeries("all_nil", []interface{}{nil, nil})
	_, err = sall.Sum()
	if err == nil {
		t.Error("Sum all nil: expected error, got nil")
	}

	// Empty returns error
	sempty := NewSeries("empty", []interface{}{})
	_, err = sempty.Sum()
	if err == nil {
		t.Error("Sum empty: expected error, got nil")
	}
}

func TestSeriesMean(t *testing.T) {
	// Basic mean
	s := NewSeries("nums", []interface{}{2, 4, 6})
	mean, err := s.Mean()
	if err != nil {
		t.Fatalf("Mean: unexpected error: %v", err)
	}
	if mean != 4.0 {
		t.Errorf("Mean: expected 4.0, got %v", mean)
	}

	// With nils
	sn := NewSeries("with_nils", []interface{}{nil, 3, nil, 9})
	mean, err = sn.Mean()
	if err != nil {
		t.Fatalf("Mean with nils: unexpected error: %v", err)
	}
	if mean != 6.0 {
		t.Errorf("Mean with nils: expected 6.0, got %v", mean)
	}
}

func TestSeriesCount(t *testing.T) {
	// Non-nil count
	s := NewSeries("nums", []interface{}{1, 2, 3, 4, 5})
	if s.Count() != 5 {
		t.Errorf("Count: expected 5, got %d", s.Count())
	}

	// All nil → 0
	sall := NewSeries("all_nil", []interface{}{nil, nil, nil})
	if sall.Count() != 0 {
		t.Errorf("Count all nil: expected 0, got %d", sall.Count())
	}

	// With nils
	sn := NewSeries("with_nils", []interface{}{1, nil, 3, nil, 5})
	if sn.Count() != 3 {
		t.Errorf("Count with nils: expected 3, got %d", sn.Count())
	}
}
