package gopandas

// DropNA returns a new DataFrame with rows containing any nil value removed.
func (df *DataFrame) DropNA() *DataFrame {
	result := NewDataFrame(df.columns)
	for i, row := range df.data {
		hasNull := false
		for _, val := range row {
			if val == nil {
				hasNull = true
				break
			}
		}
		if !hasNull {
			result.data = append(result.data, row)
			result.index = append(result.index, df.index[i])
		}
	}
	return result
}

// FillNA returns a new DataFrame with nil values replaced by value.
func (df *DataFrame) FillNA(value interface{}) *DataFrame {
	result := NewDataFrame(df.columns)
	for i, row := range df.data {
		newRow := make([]interface{}, len(row))
		for j, val := range row {
			if val == nil {
				newRow[j] = value
			} else {
				newRow[j] = val
			}
		}
		result.data = append(result.data, newRow)
		result.index = append(result.index, df.index[i])
	}
	return result
}

// HasNA returns true if any cell in the DataFrame is nil.
func (df *DataFrame) HasNA() bool {
	for _, row := range df.data {
		for _, val := range row {
			if val == nil {
				return true
			}
		}
	}
	return false
}

// IsNull returns a bool slice where true means the value at that position is nil.
func (s *Series) IsNull() []bool {
	result := make([]bool, len(s.data))
	for i, val := range s.data {
		result[i] = val == nil
	}
	return result
}

// NotNull returns a bool slice where true means the value at that position is not nil.
func (s *Series) NotNull() []bool {
	result := make([]bool, len(s.data))
	for i, val := range s.data {
		result[i] = val != nil
	}
	return result
}

// DropNA returns a new Series with nil values removed.
func (s *Series) DropNA() *Series {
	newData := make([]interface{}, 0, len(s.data))
	for _, val := range s.data {
		if val != nil {
			newData = append(newData, val)
		}
	}
	return NewSeries(s.name, newData)
}

// FillNA returns a new Series with nil values replaced by value.
func (s *Series) FillNA(value interface{}) *Series {
	newData := make([]interface{}, len(s.data))
	for i, val := range s.data {
		if val == nil {
			newData[i] = value
		} else {
			newData[i] = val
		}
	}
	return NewSeries(s.name, newData)
}
