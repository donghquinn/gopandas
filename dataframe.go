package gopandas

import (
	"fmt"
	"reflect"
)

type DataFrame struct {
	columns []string
	data    [][]interface{}
	index   []interface{}
}

type Series struct {
	name   string
	data   []interface{}
	dtype  reflect.Type
	index  []interface{}
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
	
	for i := 0; i < len(df.columns)*15; i++ {
		result += "-"
	}
	result += "\n"
	
	for _, row := range df.data {
		for _, val := range row {
			result += fmt.Sprintf("%-15v", val)
		}
		result += "\n"
	}
	
	return result
}