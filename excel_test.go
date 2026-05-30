package gopandas

import (
	"archive/zip"
	"bytes"
	"os"
	"testing"
)

// createTestXLSX creates a minimal valid .xlsx file in a temp path and returns
// the file path and a cleanup function.
func createTestXLSX(t *testing.T) (string, func()) {
	t.Helper()

	sheetXML := `<?xml version="1.0" encoding="UTF-8"?>
<worksheet>
  <sheetData>
    <row><c t="inlineStr"><is><t>name</t></is></c><c t="inlineStr"><is><t>age</t></is></c></row>
    <row><c t="inlineStr"><is><t>Alice</t></is></c><c><v>25</v></c></row>
    <row><c t="inlineStr"><is><t>Bob</t></is></c><c><v>30</v></c></row>
  </sheetData>
</worksheet>`

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	fw, err := w.Create("xl/worksheets/sheet1.xml")
	if err != nil {
		t.Fatalf("createTestXLSX: create zip entry: %v", err)
	}
	_, err = fw.Write([]byte(sheetXML))
	if err != nil {
		t.Fatalf("createTestXLSX: write sheet xml: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("createTestXLSX: close zip writer: %v", err)
	}

	f, err := os.CreateTemp("", "test_*.xlsx")
	if err != nil {
		t.Fatalf("createTestXLSX: CreateTemp: %v", err)
	}
	if _, err := f.Write(buf.Bytes()); err != nil {
		f.Close()
		os.Remove(f.Name())
		t.Fatalf("createTestXLSX: write temp file: %v", err)
	}
	f.Close()

	path := f.Name()
	cleanup := func() { os.Remove(path) }
	return path, cleanup
}

func TestReadExcelXLSX(t *testing.T) {
	path, cleanup := createTestXLSX(t)
	defer cleanup()

	df, err := ReadExcel(path)
	if err != nil {
		t.Fatalf("ReadExcelXLSX: unexpected error: %v", err)
	}

	rows, cols := df.Shape()
	if rows != 2 {
		t.Errorf("ReadExcelXLSX: expected 2 rows, got %d", rows)
	}
	if cols != 2 {
		t.Errorf("ReadExcelXLSX: expected 2 cols, got %d", cols)
	}

	columns := df.Columns()
	if columns[0] != "name" {
		t.Errorf("ReadExcelXLSX: col[0] expected name, got %q", columns[0])
	}
	if columns[1] != "age" {
		t.Errorf("ReadExcelXLSX: col[1] expected age, got %q", columns[1])
	}

	ageCol, err := df.GetColumn("age")
	if err != nil {
		t.Fatalf("ReadExcelXLSX: GetColumn age: %v", err)
	}
	ages := ageCol.Values()
	if ages[0].(int) != 25 {
		t.Errorf("ReadExcelXLSX: age[0] expected 25, got %v", ages[0])
	}
	if ages[1].(int) != 30 {
		t.Errorf("ReadExcelXLSX: age[1] expected 30, got %v", ages[1])
	}
}

func TestReadExcelUnsupportedFormat(t *testing.T) {
	_, err := ReadExcel("testfile.txt")
	if err == nil {
		t.Error("ReadExcelUnsupportedFormat: expected error for .txt file, got nil")
	}
}

func TestReadExcelNonExistent(t *testing.T) {
	_, err := ReadExcel("/nonexistent/path/to/file.xlsx")
	if err == nil {
		t.Error("ReadExcelNonExistent: expected error for nonexistent file, got nil")
	}
}

// ── Tests reading static testdata files ──────────────────────────────────────

func TestReadExcelBasicFile(t *testing.T) {
	df, err := ReadExcel("testdata/basic.xlsx")
	if err != nil {
		t.Fatalf("ReadExcelBasicFile: %v", err)
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
	cityCol, _ := df.GetColumn("city")
	if cityCol.Values()[0] != "New York" {
		t.Errorf("row 0 city: expected New York, got %v", cityCol.Values()[0])
	}
}

func TestReadExcelNullsFile(t *testing.T) {
	// nulls.xlsx: name / score / grade, 4 rows
	// Alice, nil, A
	// Bob, 85, nil
	// Charlie, 90, B
	// nil, 70, C
	df, err := ReadExcel("testdata/nulls.xlsx")
	if err != nil {
		t.Fatalf("ReadExcelNullsFile: %v", err)
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

func TestReadExcelXLSFile(t *testing.T) {
	df, err := ReadExcel("testdata/basic.xls")
	if err != nil {
		t.Fatalf("ReadExcelXLSFile: %v", err)
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
	ageCol, _ := df.GetColumn("age")
	if ageCol.Values()[0].(int) != 25 {
		t.Errorf("row 0 age: expected 25, got %v", ageCol.Values()[0])
	}
	nameCol, _ := df.GetColumn("name")
	if nameCol.Values()[4] != "Eve" {
		t.Errorf("row 4 name: expected Eve, got %v", nameCol.Values()[4])
	}
}
