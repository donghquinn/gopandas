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
		found := false
		for j, dfCol := range df.columns {
			if dfCol == col {
				colIndices[i] = j
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("column '%s' not found", col)
		}
	}
	
	result := NewDataFrame(columns)
	
	for i, row := range df.data {
		newRow := make([]interface{}, len(columns))
		for j, colIdx := range colIndices {
			newRow[j] = row[colIdx]
		}
		result.data = append(result.data, newRow)
		result.index = append(result.index, df.index[i])
	}
	
	return result, nil
}

func (df *DataFrame) Sort(column string, ascending bool) (*DataFrame, error) {
	colIndex := -1
	for i, col := range df.columns {
		if col == column {
			colIndex = i
			break
		}
	}
	
	if colIndex == -1 {
		return nil, fmt.Errorf("column '%s' not found", column)
	}
	
	result := NewDataFrame(df.columns)
	result.data = make([][]interface{}, len(df.data))
	result.index = make([]interface{}, len(df.index))
	
	copy(result.data, df.data)
	copy(result.index, df.index)
	
	sort.Slice(result.data, func(i, j int) bool {
		valI := result.data[i][colIndex]
		valJ := result.data[j][colIndex]
		
		comp := compareValues(valI, valJ)
		if ascending {
			return comp < 0
		}
		return comp > 0
	})
	
	return result, nil
}

func (df *DataFrame) GroupBy(column string) (map[interface{}]*DataFrame, error) {
	colIndex := -1
	for i, col := range df.columns {
		if col == column {
			colIndex = i
			break
		}
	}
	
	if colIndex == -1 {
		return nil, fmt.Errorf("column '%s' not found", column)
	}
	
	groups := make(map[interface{}]*DataFrame)
	
	for i, row := range df.data {
		key := row[colIndex]
		
		if groups[key] == nil {
			groups[key] = NewDataFrame(df.columns)
		}
		
		groups[key].data = append(groups[key].data, row)
		groups[key].index = append(groups[key].index, df.index[i])
	}
	
	return groups, nil
}

func (s *Series) Sum() (interface{}, error) {
	if len(s.data) == 0 {
		return nil, fmt.Errorf("series is empty")
	}
	
	var sum float64
	count := 0
	
	for _, val := range s.data {
		if val != nil {
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
	}
	
	if count == 0 {
		return nil, fmt.Errorf("no numeric values found")
	}
	
	return sum, nil
}

func (s *Series) Mean() (float64, error) {
	sum, err := s.Sum()
	if err != nil {
		return 0, err
	}
	
	count := 0
	for _, val := range s.data {
		if val != nil {
			switch val.(type) {
			case int, float64, float32:
				count++
			}
		}
	}
	
	if count == 0 {
		return 0, fmt.Errorf("no numeric values found")
	}
	
	return sum.(float64) / float64(count), nil
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
	
	switch va := a.(type) {
	case int:
		if vb, ok := b.(int); ok {
			if va < vb {
				return -1
			} else if va > vb {
				return 1
			}
			return 0
		}
	case float64:
		if vb, ok := b.(float64); ok {
			if va < vb {
				return -1
			} else if va > vb {
				return 1
			}
			return 0
		}
	case string:
		if vb, ok := b.(string); ok {
			if va < vb {
				return -1
			} else if va > vb {
				return 1
			}
			return 0
		}
	}
	
	return 0
}