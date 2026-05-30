package gopandas

import (
	"fmt"
	"os"
	"testing"
)

// ── Tests reading static testdata files ──────────────────────────────────────

func TestReadCSVFromFile(t *testing.T) {
	df, err := ReadCSV("testdata/basic.csv")
	if err != nil {
		t.Fatalf("ReadCSVFromFile: %v", err)
	}
	rows, cols := df.Shape()
	if rows != 5 || cols != 3 {
		t.Errorf("expected (5,3), got (%d,%d)", rows, cols)
	}
	for i, want := range []string{"name", "age", "city"} {
		if got := df.Columns()[i]; got != want {
			t.Errorf("col[%d]: expected %q, got %q", i, want, got)
		}
	}
	nameCol, _ := df.GetColumn("name")
	if nameCol.Values()[0] != "Alice" {
		t.Errorf("row 0 name: expected Alice, got %v", nameCol.Values()[0])
	}
	ageCol, _ := df.GetColumn("age")
	if ageCol.Values()[0].(int) != 25 {
		t.Errorf("row 0 age: expected 25, got %v", ageCol.Values()[0])
	}
}

func TestReadCSVTypesFromFile(t *testing.T) {
	df, err := ReadCSV("testdata/types.csv")
	if err != nil {
		t.Fatalf("ReadCSVTypesFromFile: %v", err)
	}
	rows, cols := df.Shape()
	if rows != 2 || cols != 5 {
		t.Errorf("expected (2,5), got (%d,%d)", rows, cols)
	}
	intCol, _ := df.GetColumn("int_col")
	floatCol, _ := df.GetColumn("float_col")
	boolCol, _ := df.GetColumn("bool_col")
	strCol, _ := df.GetColumn("str_col")
	emptyCol, _ := df.GetColumn("empty_col")

	if _, ok := intCol.Values()[0].(int); !ok {
		t.Errorf("int_col expected int, got %T", intCol.Values()[0])
	}
	if _, ok := floatCol.Values()[0].(float64); !ok {
		t.Errorf("float_col expected float64, got %T", floatCol.Values()[0])
	}
	if _, ok := boolCol.Values()[0].(bool); !ok {
		t.Errorf("bool_col expected bool, got %T", boolCol.Values()[0])
	}
	if _, ok := strCol.Values()[0].(string); !ok {
		t.Errorf("str_col expected string, got %T", strCol.Values()[0])
	}
	if emptyCol.Values()[0] != nil {
		t.Errorf("empty_col row 0: expected nil, got %v", emptyCol.Values()[0])
	}
	// second row: negative int, float, bool false, string, nil
	if intCol.Values()[1].(int) != -10 {
		t.Errorf("int_col row 1: expected -10, got %v", intCol.Values()[1])
	}
	if boolCol.Values()[1].(bool) != false {
		t.Errorf("bool_col row 1: expected false, got %v", boolCol.Values()[1])
	}
}

func TestReadCSVNoHeaderFromFile(t *testing.T) {
	df, err := ReadCSV("testdata/no_header.csv", WithHeader(false))
	if err != nil {
		t.Fatalf("ReadCSVNoHeaderFromFile: %v", err)
	}
	rows, cols := df.Shape()
	if rows != 3 || cols != 3 {
		t.Errorf("expected (3,3), got (%d,%d)", rows, cols)
	}
	for i, col := range df.Columns() {
		want := fmt.Sprintf("col_%d", i)
		if col != want {
			t.Errorf("col[%d]: expected %q, got %q", i, want, col)
		}
	}
}

func TestReadCSVTabDelimFromFile(t *testing.T) {
	df, err := ReadCSV("testdata/tab.tsv", WithDelimiter('\t'))
	if err != nil {
		t.Fatalf("ReadCSVTabDelimFromFile: %v", err)
	}
	rows, cols := df.Shape()
	if rows != 3 || cols != 3 {
		t.Errorf("expected (3,3), got (%d,%d)", rows, cols)
	}
}

func TestReadCSVNullsFromFile(t *testing.T) {
	// nulls.csv: name,score,grade
	// Alice,,A   → score=nil
	// Bob,85,    → grade=nil
	// Charlie,90,B
	// ,70,C      → name=nil
	df, err := ReadCSV("testdata/nulls.csv")
	if err != nil {
		t.Fatalf("ReadCSVNullsFromFile: %v", err)
	}
	rows, cols := df.Shape()
	if rows != 4 || cols != 3 {
		t.Errorf("expected (4,3), got (%d,%d)", rows, cols)
	}
	if !df.HasNA() {
		t.Error("expected HasNA=true, got false")
	}
	scoreCol, _ := df.GetColumn("score")
	if scoreCol.Values()[0] != nil {
		t.Errorf("Alice score: expected nil, got %v", scoreCol.Values()[0])
	}
	if scoreCol.Values()[1].(int) != 85 {
		t.Errorf("Bob score: expected 85, got %v", scoreCol.Values()[1])
	}
	gradeCol, _ := df.GetColumn("grade")
	if gradeCol.Values()[1] != nil {
		t.Errorf("Bob grade: expected nil, got %v", gradeCol.Values()[1])
	}
	nameCol, _ := df.GetColumn("name")
	if nameCol.Values()[3] != nil {
		t.Errorf("row 3 name: expected nil, got %v", nameCol.Values()[3])
	}
}

func TestReadCSVBasic(t *testing.T) {
	content := "name,age,city\nAlice,25,New York\nBob,30,London\nCharlie,35,Paris\n"
	f, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	df, err := ReadCSV(f.Name())
	if err != nil {
		t.Fatalf("ReadCSV basic: unexpected error: %v", err)
	}

	rows, cols := df.Shape()
	if rows != 3 || cols != 3 {
		t.Errorf("ReadCSV basic: expected (3,3), got (%d,%d)", rows, cols)
	}

	columns := df.Columns()
	expected := []string{"name", "age", "city"}
	for i, c := range expected {
		if columns[i] != c {
			t.Errorf("ReadCSV basic col[%d]: expected %q, got %q", i, c, columns[i])
		}
	}
}

func TestReadCSVTypeInference(t *testing.T) {
	// int, float, bool, string, empty
	content := "i,f,b,s,e\n42,3.14,true,hello,\n"
	f, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	df, err := ReadCSV(f.Name())
	if err != nil {
		t.Fatalf("ReadCSV type inference: unexpected error: %v", err)
	}

	colI, _ := df.GetColumn("i")
	colF, _ := df.GetColumn("f")
	colB, _ := df.GetColumn("b")
	colS, _ := df.GetColumn("s")
	colE, _ := df.GetColumn("e")

	if _, ok := colI.Values()[0].(int); !ok {
		t.Errorf("type inference: i expected int, got %T", colI.Values()[0])
	}
	if _, ok := colF.Values()[0].(float64); !ok {
		t.Errorf("type inference: f expected float64, got %T", colF.Values()[0])
	}
	if _, ok := colB.Values()[0].(bool); !ok {
		t.Errorf("type inference: b expected bool, got %T", colB.Values()[0])
	}
	if _, ok := colS.Values()[0].(string); !ok {
		t.Errorf("type inference: s expected string, got %T", colS.Values()[0])
	}
	if colE.Values()[0] != nil {
		t.Errorf("type inference: empty cell expected nil, got %v", colE.Values()[0])
	}
}

func TestReadCSVNoHeader(t *testing.T) {
	content := "Alice,25,New York\nBob,30,London\n"
	f, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	df, err := ReadCSV(f.Name(), WithHeader(false))
	if err != nil {
		t.Fatalf("ReadCSV no header: unexpected error: %v", err)
	}

	rows, cols := df.Shape()
	if rows != 2 || cols != 3 {
		t.Errorf("ReadCSV no header: expected (2,3), got (%d,%d)", rows, cols)
	}

	columns := df.Columns()
	for i, c := range columns {
		expected := fmt.Sprintf("col_%d", i)
		if c != expected {
			t.Errorf("ReadCSV no header col[%d]: expected %q, got %q", i, expected, c)
		}
	}
}

func TestReadCSVCustomDelimiter(t *testing.T) {
	content := "name\tage\tcity\nAlice\t25\tNew York\nBob\t30\tLondon\n"
	f, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	df, err := ReadCSV(f.Name(), WithDelimiter('\t'))
	if err != nil {
		t.Fatalf("ReadCSV custom delimiter: unexpected error: %v", err)
	}

	rows, cols := df.Shape()
	if rows != 2 || cols != 3 {
		t.Errorf("ReadCSV tab delimiter: expected (2,3), got (%d,%d)", rows, cols)
	}
}

func TestReadCSVNonExistent(t *testing.T) {
	_, err := ReadCSV("/nonexistent/path/to/file.csv")
	if err == nil {
		t.Error("ReadCSV nonexistent: expected error, got nil")
	}
}

func TestToCSV(t *testing.T) {
	df := makePersonDF()

	f, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	f.Close()

	err = df.ToCSV(f.Name())
	if err != nil {
		t.Fatalf("ToCSV: unexpected error: %v", err)
	}

	// Read back and verify
	df2, err := ReadCSV(f.Name())
	if err != nil {
		t.Fatalf("ToCSV read back: unexpected error: %v", err)
	}

	rows, cols := df2.Shape()
	if rows != 3 || cols != 3 {
		t.Errorf("ToCSV read back: expected (3,3), got (%d,%d)", rows, cols)
	}

	nameCol, _ := df2.GetColumn("name")
	names := nameCol.Values()
	if names[0].(string) != "Alice" {
		t.Errorf("ToCSV: name[0] expected Alice, got %v", names[0])
	}
}

func TestToCSVNoHeader(t *testing.T) {
	df := NewDataFrame([]string{"a", "b"})
	df.AddRow([]interface{}{1, 2})
	df.AddRow([]interface{}{3, 4})

	f, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	f.Close()

	err = df.ToCSV(f.Name(), WithHeader(false))
	if err != nil {
		t.Fatalf("ToCSVNoHeader: unexpected error: %v", err)
	}

	// When read back without header, the first data row is row 0
	df2, err := ReadCSV(f.Name(), WithHeader(false))
	if err != nil {
		t.Fatalf("ToCSVNoHeader read back: unexpected error: %v", err)
	}

	rows, _ := df2.Shape()
	if rows != 2 {
		t.Errorf("ToCSVNoHeader: expected 2 data rows, got %d", rows)
	}
}

func TestCSVRoundtrip(t *testing.T) {
	original := makePersonDF()

	f, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	f.Close()

	err = original.ToCSV(f.Name())
	if err != nil {
		t.Fatalf("CSVRoundtrip ToCSV: %v", err)
	}

	restored, err := ReadCSV(f.Name())
	if err != nil {
		t.Fatalf("CSVRoundtrip ReadCSV: %v", err)
	}

	origRows, origCols := original.Shape()
	restRows, restCols := restored.Shape()
	if origRows != restRows || origCols != restCols {
		t.Errorf("CSVRoundtrip shape mismatch: original (%d,%d), restored (%d,%d)",
			origRows, origCols, restRows, restCols)
	}
}
