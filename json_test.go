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
