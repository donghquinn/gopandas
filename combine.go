package gopandas

import "fmt"

// Concat concatenates multiple DataFrames vertically (row-wise).
// All DataFrames must have identical column names in the same order.
func Concat(dfs ...*DataFrame) (*DataFrame, error) {
	if len(dfs) == 0 {
		return nil, fmt.Errorf("no DataFrames provided")
	}

	base := dfs[0].columns
	for i, df := range dfs[1:] {
		if len(df.columns) != len(base) {
			return nil, fmt.Errorf("DataFrame %d column count mismatch", i+1)
		}
		for j, col := range df.columns {
			if col != base[j] {
				return nil, fmt.Errorf("DataFrame %d column %d: expected %q got %q", i+1, j, base[j], col)
			}
		}
	}

	result := NewDataFrame(base)
	for _, df := range dfs {
		for _, row := range df.data {
			result.data = append(result.data, row)
			result.index = append(result.index, len(result.index))
		}
	}
	return result, nil
}

// Merge joins two DataFrames on a common column.
// how: "inner" | "left" | "right" | "outer"
func (df *DataFrame) Merge(right *DataFrame, on string, how string) (*DataFrame, error) {
	switch how {
	case "inner", "left", "right", "outer":
	default:
		return nil, fmt.Errorf("invalid join type %q: use inner, left, right, or outer", how)
	}

	leftColIdx, err := findColIndex(df.columns, on)
	if err != nil {
		return nil, fmt.Errorf("left DataFrame: %w", err)
	}
	rightColIdx, err := findColIndex(right.columns, on)
	if err != nil {
		return nil, fmt.Errorf("right DataFrame: %w", err)
	}

	// Build result column list: all left cols + right cols (excluding join key, deduping names)
	resultCols := make([]string, len(df.columns))
	copy(resultCols, df.columns)
	leftColSet := make(map[string]bool, len(df.columns))
	for _, c := range df.columns {
		leftColSet[c] = true
	}

	rightResultIndices := make([]int, 0, len(right.columns)-1)
	for i, col := range right.columns {
		if i == rightColIdx {
			continue
		}
		name := col
		if leftColSet[col] {
			name = col + "_right"
		}
		resultCols = append(resultCols, name)
		rightResultIndices = append(rightResultIndices, i)
	}

	result := NewDataFrame(resultCols)

	// Index right rows by join key for fast lookup
	rightIndex := make(map[interface{}][]int, len(right.data))
	for i, row := range right.data {
		key := row[rightColIdx]
		rightIndex[key] = append(rightIndex[key], i)
	}

	rightMatched := make([]bool, len(right.data))

	for _, leftRow := range df.data {
		key := leftRow[leftColIdx]
		rightRows, found := rightIndex[key]

		if found {
			for _, ri := range rightRows {
				newRow := buildMergedRow(leftRow, right.data[ri], rightResultIndices, len(resultCols))
				result.data = append(result.data, newRow)
				result.index = append(result.index, len(result.index))
				rightMatched[ri] = true
			}
		} else if how == "left" || how == "outer" {
			newRow := make([]interface{}, len(resultCols))
			copy(newRow, leftRow)
			result.data = append(result.data, newRow)
			result.index = append(result.index, len(result.index))
		}
	}

	// Add unmatched right rows for right/outer joins
	if how == "right" || how == "outer" {
		for i, matched := range rightMatched {
			if matched {
				continue
			}
			rightRow := right.data[i]
			newRow := make([]interface{}, len(resultCols))
			// Set join key from right side
			newRow[leftColIdx] = rightRow[rightColIdx]
			offset := len(df.columns)
			for j, ri := range rightResultIndices {
				newRow[offset+j] = rightRow[ri]
			}
			result.data = append(result.data, newRow)
			result.index = append(result.index, len(result.index))
		}
	}

	return result, nil
}

func findColIndex(columns []string, name string) (int, error) {
	for i, col := range columns {
		if col == name {
			return i, nil
		}
	}
	return -1, fmt.Errorf("column '%s' not found", name)
}

func buildMergedRow(leftRow, rightRow []interface{}, rightResultIndices []int, totalCols int) []interface{} {
	newRow := make([]interface{}, totalCols)
	copy(newRow, leftRow)
	offset := len(leftRow)
	for j, ri := range rightResultIndices {
		newRow[offset+j] = rightRow[ri]
	}
	return newRow
}
