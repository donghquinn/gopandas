package gopandas

import (
	"fmt"
	"reflect"
	"strings"
)

type DataFrame struct {
	columns []string
	data    [][]interface{}
	index   []interface{}
}

type Series struct {
	name  string
	data  []interface{}
	dtype reflect.Type
	index []interface{}
}

func NewDataFrame(columns []string) *DataFrame {
	return &DataFrame{
		columns: columns,
		data:    make([][]interface{}, 0),
		index:   make([]interface{}, 0),
	}
}

func NewSeries(name string, data []interface{}) *Series {
	var dtype reflect.Type
	if len(data) > 0 {
		dtype = reflect.TypeOf(data[0])
	}

	index := make([]interface{}, len(data))
	for i := range index {
		index[i] = i
	}

	return &Series{
		name:  name,
		data:  data,
		dtype: dtype,
		index: index,
	}
}

func (df *DataFrame) Shape() (int, int) {
	return len(df.data), len(df.columns)
}

func (df *DataFrame) Columns() []string {
	return df.columns
}

func (df *DataFrame) Head(n int) *DataFrame {
	if n > len(df.data) {
		n = len(df.data)
	}

	result := NewDataFrame(df.columns)
	result.data = df.data[:n]
	result.index = df.index[:n]

	return result
}

func (df *DataFrame) AddRow(row []interface{}) error {
	if len(row) != len(df.columns) {
		return fmt.Errorf("row length %d does not match columns length %d", len(row), len(df.columns))
	}

	df.data = append(df.data, row)
	df.index = append(df.index, len(df.data)-1)

	return nil
}

func (df *DataFrame) GetColumn(name string) (*Series, error) {
	colIndex := -1
	for i, col := range df.columns {
		if col == name {
			colIndex = i
			break
		}
	}

	if colIndex == -1 {
		return nil, fmt.Errorf("column '%s' not found", name)
	}

	columnData := make([]interface{}, len(df.data))
	for i, row := range df.data {
		columnData[i] = row[colIndex]
	}

	return NewSeries(name, columnData), nil
}

func (df *DataFrame) String() string {
	result := ""

	for _, col := range df.columns {
		result += fmt.Sprintf("%-15s", col)
	}
	result += "\n"

	result += strings.Repeat("-", len(df.columns)*15) + "\n"

	for _, row := range df.data {
		for _, val := range row {
			result += fmt.Sprintf("%-15v", val)
		}
		result += "\n"
	}

	return result
}

func (df *DataFrame) Tail(n int) *DataFrame {
	rows := len(df.data)
	if n > rows {
		n = rows
	}
	result := NewDataFrame(df.columns)
	result.data = df.data[rows-n:]
	result.index = df.index[rows-n:]
	return result
}

func (df *DataFrame) Rename(mapping map[string]string) *DataFrame {
	newCols := make([]string, len(df.columns))
	for i, col := range df.columns {
		if newName, ok := mapping[col]; ok {
			newCols[i] = newName
		} else {
			newCols[i] = col
		}
	}
	return &DataFrame{columns: newCols, data: df.data, index: df.index}
}

func (df *DataFrame) Drop(columns ...string) (*DataFrame, error) {
	dropSet := make(map[string]bool)
	for _, col := range columns {
		found := false
		for _, dfCol := range df.columns {
			if dfCol == col {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("column '%s' not found", col)
		}
		dropSet[col] = true
	}

	newCols := make([]string, 0, len(df.columns))
	keepIndices := make([]int, 0, len(df.columns))
	for i, col := range df.columns {
		if !dropSet[col] {
			newCols = append(newCols, col)
			keepIndices = append(keepIndices, i)
		}
	}

	result := NewDataFrame(newCols)
	for i, row := range df.data {
		newRow := make([]interface{}, len(keepIndices))
		for j, idx := range keepIndices {
			newRow[j] = row[idx]
		}
		result.data = append(result.data, newRow)
		result.index = append(result.index, df.index[i])
	}
	return result, nil
}

// Iloc returns rows by integer position range [start, end).
// Negative indices count from the end.
func (df *DataFrame) Iloc(start, end int) (*DataFrame, error) {
	rows := len(df.data)
	if start < 0 {
		start = rows + start
	}
	if end < 0 {
		end = rows + end
	}
	if start < 0 || start > rows {
		return nil, fmt.Errorf("start index %d out of range", start)
	}
	if end < 0 || end > rows {
		return nil, fmt.Errorf("end index %d out of range", end)
	}
	if start > end {
		return nil, fmt.Errorf("start index must be <= end index")
	}
	result := NewDataFrame(df.columns)
	result.data = df.data[start:end]
	result.index = df.index[start:end]
	return result, nil
}

// IlocRows returns rows at the given integer positions.
func (df *DataFrame) IlocRows(indices ...int) (*DataFrame, error) {
	rows := len(df.data)
	result := NewDataFrame(df.columns)
	for _, idx := range indices {
		if idx < 0 || idx >= rows {
			return nil, fmt.Errorf("index %d out of range", idx)
		}
		result.data = append(result.data, df.data[idx])
		result.index = append(result.index, df.index[idx])
	}
	return result, nil
}

// Apply applies fn to each row and returns a new DataFrame with the results.
func (df *DataFrame) Apply(fn func(row []interface{}) []interface{}) *DataFrame {
	result := NewDataFrame(df.columns)
	for i, row := range df.data {
		result.data = append(result.data, fn(row))
		result.index = append(result.index, df.index[i])
	}
	return result
}

// Dtypes returns a map of column name to Go type name, inferred from data.
func (df *DataFrame) Dtypes() map[string]string {
	result := make(map[string]string, len(df.columns))
	for i, col := range df.columns {
		typeName := "unknown"
		for _, row := range df.data {
			if row[i] != nil {
				typeName = fmt.Sprintf("%T", row[i])
				break
			}
		}
		result[col] = typeName
	}
	return result
}

// Index returns the row index values.
func (df *DataFrame) Index() []interface{} {
	return df.index
}

// SetColumn sets or replaces a column with the given Series data.
func (df *DataFrame) SetColumn(name string, values []interface{}) error {
	if len(values) != len(df.data) && len(df.data) > 0 {
		return fmt.Errorf("values length %d does not match row count %d", len(values), len(df.data))
	}

	for i, col := range df.columns {
		if col == name {
			for j, row := range df.data {
				row[i] = values[j]
			}
			return nil
		}
	}

	// New column
	df.columns = append(df.columns, name)
	for i, row := range df.data {
		if i < len(values) {
			df.data[i] = append(row, values[i])
		} else {
			df.data[i] = append(row, nil)
		}
	}
	return nil
}
