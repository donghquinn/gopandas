package gopandas

import (
	"fmt"
	"sort"
)

func (df *DataFrame) Filter(predicate func(row []interface{}) bool) *DataFrame {
	result := NewDataFrame(df.columns)

	for i, row := range df.data {
		if predicate(row) {
			result.data = append(result.data, row)
			result.index = append(result.index, df.index[i])
		}
	}

	return result
}

func (df *DataFrame) Select(columns ...string) (*DataFrame, error) {
	colIndices := make([]int, len(columns))

	for i, col := range columns {
		idx, err := findColIndex(df.columns, col)
		if err != nil {
			return nil, err
		}
		colIndices[i] = idx
	}

	result := NewDataFrame(columns)
	result.data = make([][]interface{}, len(df.data))
	result.index = make([]interface{}, len(df.data))

	for i, row := range df.data {
		newRow := make([]interface{}, len(colIndices))
		for j, colIdx := range colIndices {
			newRow[j] = row[colIdx]
		}
		result.data[i] = newRow
		result.index[i] = df.index[i]
	}

	return result, nil
}

func (df *DataFrame) Sort(column string, ascending bool) (*DataFrame, error) {
	colIndex, err := findColIndex(df.columns, column)
	if err != nil {
		return nil, err
	}

	// Sort a permutation so data and index stay aligned.
	perm := make([]int, len(df.data))
	keys := make([]interface{}, len(df.data))
	for i, row := range df.data {
		perm[i] = i
		keys[i] = row[colIndex]
	}

	sort.Slice(perm, func(i, j int) bool {
		comp := compareValues(keys[perm[i]], keys[perm[j]])
		if ascending {
			return comp < 0
		}
		return comp > 0
	})

	result := NewDataFrame(df.columns)
	result.data = make([][]interface{}, len(df.data))
	result.index = make([]interface{}, len(df.index))
	for i, p := range perm {
		result.data[i] = df.data[p]
		result.index[i] = df.index[p]
	}

	return result, nil
}

func (df *DataFrame) GroupBy(column string) (map[interface{}]*DataFrame, error) {
	colIndex, err := findColIndex(df.columns, column)
	if err != nil {
		return nil, err
	}

	groups := make(map[interface{}]*DataFrame)

	for i, row := range df.data {
		key := row[colIndex]

		group := groups[key]
		if group == nil {
			group = NewDataFrame(df.columns)
			groups[key] = group
		}

		group.data = append(group.data, row)
		group.index = append(group.index, df.index[i])
	}

	return groups, nil
}

// sumCount returns the sum and count of numeric (int, float32, float64) values in one pass.
func (s *Series) sumCount() (float64, int) {
	var sum float64
	count := 0

	for _, val := range s.data {
		switch v := val.(type) {
		case int:
			sum += float64(v)
			count++
		case float64:
			sum += v
			count++
		case float32:
			sum += float64(v)
			count++
		}
	}

	return sum, count
}

func (s *Series) Sum() (interface{}, error) {
	if len(s.data) == 0 {
		return nil, fmt.Errorf("series is empty")
	}

	sum, count := s.sumCount()
	if count == 0 {
		return nil, fmt.Errorf("no numeric values found")
	}

	return sum, nil
}

func (s *Series) Mean() (float64, error) {
	if len(s.data) == 0 {
		return 0, fmt.Errorf("series is empty")
	}

	sum, count := s.sumCount()
	if count == 0 {
		return 0, fmt.Errorf("no numeric values found")
	}

	return sum / float64(count), nil
}

func (s *Series) Count() int {
	count := 0
	for _, val := range s.data {
		if val != nil {
			count++
		}
	}
	return count
}

func compareValues(a, b interface{}) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// Numeric values compare across int/float32/float64.
	fa, aNum := toFloat64(a)
	fb, bNum := toFloat64(b)
	if aNum && bNum {
		switch {
		case fa < fb:
			return -1
		case fa > fb:
			return 1
		}
		return 0
	}

	if va, ok := a.(string); ok {
		if vb, ok := b.(string); ok {
			switch {
			case va < vb:
				return -1
			case va > vb:
				return 1
			}
			return 0
		}
	}

	return 0
}

// toFloat64 converts numeric values (int, float32, float64) to float64.
func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case float64:
		return n, true
	case float32:
		return float64(n), true
	}
	return 0, false
}
