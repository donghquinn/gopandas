package gopandas

import (
	"encoding/json"
	"fmt"
	"os"
)

// ReadJSON reads a JSON file containing an array of objects into a DataFrame.
// Keys from the first object set the column order; subsequent objects may add more columns.
func ReadJSON(filename string) (*DataFrame, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var records []map[string]interface{}
	if err := json.Unmarshal(raw, &records); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if len(records) == 0 {
		return NewDataFrame([]string{}), nil
	}

	// Collect columns preserving insertion order
	colSet := make(map[string]bool)
	cols := make([]string, 0)
	for _, rec := range records {
		for k := range rec {
			if !colSet[k] {
				cols = append(cols, k)
				colSet[k] = true
			}
		}
	}

	df := NewDataFrame(cols)
	for _, rec := range records {
		row := make([]interface{}, len(cols))
		for i, col := range cols {
			val, ok := rec[col]
			if ok {
				row[i] = normalizeJSONValue(val)
			}
		}
		df.AddRow(row)
	}
	return df, nil
}

// ToJSON writes the DataFrame to a JSON file as an array of objects.
func (df *DataFrame) ToJSON(filename string) error {
	records := make([]map[string]interface{}, len(df.data))
	for i, row := range df.data {
		rec := make(map[string]interface{}, len(df.columns))
		for j, col := range df.columns {
			rec[col] = row[j]
		}
		records[i] = rec
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// normalizeJSONValue converts JSON number values to int where possible.
func normalizeJSONValue(val interface{}) interface{} {
	if f, ok := val.(float64); ok {
		if f == float64(int64(f)) {
			return int(f)
		}
		return f
	}
	return val
}
