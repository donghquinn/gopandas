package gopandas

import (
	"fmt"
	"os"
	"testing"
)

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
