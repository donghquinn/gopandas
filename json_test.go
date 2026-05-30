package gopandas

import (
	"os"
	"testing"
)

func TestReadJSONBasic(t *testing.T) {
	content := `[{"name":"Alice","age":25},{"name":"Bob","age":30},{"name":"Charlie","age":35}]`
	f, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	df, err := ReadJSON(f.Name())
	if err != nil {
		t.Fatalf("ReadJSONBasic: unexpected error: %v", err)
	}

	rows, cols := df.Shape()
	if rows != 3 {
		t.Errorf("ReadJSONBasic: expected 3 rows, got %d", rows)
	}
	if cols != 2 {
		t.Errorf("ReadJSONBasic: expected 2 cols, got %d", cols)
	}
}

func TestReadJSONEmptyArray(t *testing.T) {
	content := `[]`
	f, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	df, err := ReadJSON(f.Name())
	if err != nil {
		t.Fatalf("ReadJSONEmptyArray: unexpected error: %v", err)
	}

	rows, cols := df.Shape()
	if rows != 0 {
		t.Errorf("ReadJSONEmptyArray: expected 0 rows, got %d", rows)
	}
	if cols != 0 {
		t.Errorf("ReadJSONEmptyArray: expected 0 cols, got %d", cols)
	}
}

func TestReadJSONNonExistent(t *testing.T) {
	_, err := ReadJSON("/nonexistent/path/to/file.json")
	if err == nil {
		t.Error("ReadJSONNonExistent: expected error, got nil")
	}
}

func TestReadJSONInvalidSyntax(t *testing.T) {
	content := `[{"name":"Alice", "age": }]` // malformed JSON
	f, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	_, err = ReadJSON(f.Name())
	if err == nil {
		t.Error("ReadJSONInvalidSyntax: expected error for malformed JSON, got nil")
	}
}

func TestToJSONBasic(t *testing.T) {
	df := makePersonDF()

	f, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	f.Close()

	err = df.ToJSON(f.Name())
	if err != nil {
		t.Fatalf("ToJSONBasic: unexpected error: %v", err)
	}

	// Read the file back and check it's valid JSON
	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("ToJSONBasic: read file: %v", err)
	}
	if len(data) == 0 {
		t.Error("ToJSONBasic: expected non-empty JSON output")
	}
	// Basic check: starts with '[' (array)
	if data[0] != '[' {
		t.Errorf("ToJSONBasic: expected JSON array starting with '[', got %q", string(data[:1]))
	}
}

func TestJSONRoundtrip(t *testing.T) {
	original := makePersonDF()

	f, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	f.Close()

	err = original.ToJSON(f.Name())
	if err != nil {
		t.Fatalf("JSONRoundtrip ToJSON: %v", err)
	}

	restored, err := ReadJSON(f.Name())
	if err != nil {
		t.Fatalf("JSONRoundtrip ReadJSON: %v", err)
	}

	origRows, origCols := original.Shape()
	restRows, restCols := restored.Shape()
	if origRows != restRows || origCols != restCols {
		t.Errorf("JSONRoundtrip shape mismatch: original (%d,%d), restored (%d,%d)",
			origRows, origCols, restRows, restCols)
	}
}

func TestJSONNullValues(t *testing.T) {
	// Build a DataFrame with nil values
	df := NewDataFrame([]string{"name", "score"})
	df.AddRow([]interface{}{"Alice", nil})
	df.AddRow([]interface{}{"Bob", 95})

	f, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	f.Close()

	err = df.ToJSON(f.Name())
	if err != nil {
		t.Fatalf("JSONNullValues ToJSON: %v", err)
	}

	// Read back
	df2, err := ReadJSON(f.Name())
	if err != nil {
		t.Fatalf("JSONNullValues ReadJSON: %v", err)
	}

	rows, _ := df2.Shape()
	if rows != 2 {
		t.Fatalf("JSONNullValues: expected 2 rows, got %d", rows)
	}

	scoreCol, err := df2.GetColumn("score")
	if err != nil {
		t.Fatalf("JSONNullValues: GetColumn score: %v", err)
	}
	scores := scoreCol.Values()
	// Alice's score was nil → should be nil after roundtrip
	if scores[0] != nil {
		t.Errorf("JSONNullValues: score[0] expected nil, got %v", scores[0])
	}
}

// ── Tests reading static testdata files ──────────────────────────────────────

// TestReadJSONFloatValues covers normalizeJSONValue's "return f" branch
// (non-integer float64 values stay as float64 instead of being converted to int).
func TestReadJSONFloatValues(t *testing.T) {
	content := `[{"item":"A","price":9.99},{"item":"B","price":4.50}]`
	f, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	df, err := ReadJSON(f.Name())
	if err != nil {
		t.Fatalf("ReadJSONFloatValues: %v", err)
	}
	priceCol, _ := df.GetColumn("price")
	if _, ok := priceCol.Values()[0].(float64); !ok {
		t.Errorf("price should be float64, got %T", priceCol.Values()[0])
	}
	if priceCol.Values()[0].(float64) != 9.99 {
		t.Errorf("price[0]: expected 9.99, got %v", priceCol.Values()[0])
	}
}

func TestReadJSONFromFile(t *testing.T) {
	df, err := ReadJSON("testdata/basic.json")
	if err != nil {
		t.Fatalf("ReadJSONFromFile: %v", err)
	}
	rows, cols := df.Shape()
	if rows != 5 || cols != 3 {
		t.Errorf("expected (5,3), got (%d,%d)", rows, cols)
	}
	// Column order from JSON maps is non-deterministic; verify by name.
	for _, want := range []string{"name", "age", "city"} {
		if _, err := df.GetColumn(want); err != nil {
			t.Errorf("missing column %q", want)
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

func TestReadJSONNullsFromFile(t *testing.T) {
	// null_values.json: name / score / grade
	// Alice, null, A
	// Bob, 85, null
	// Charlie, 90, B
	df, err := ReadJSON("testdata/null_values.json")
	if err != nil {
		t.Fatalf("ReadJSONNullsFromFile: %v", err)
	}
	rows, cols := df.Shape()
	if rows != 3 || cols != 3 {
		t.Errorf("expected (3,3), got (%d,%d)", rows, cols)
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
}
