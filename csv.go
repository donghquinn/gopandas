package gopandas

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ReadCSV(filename string, options ...CSVOption) (*DataFrame, error) {
	config := &CSVConfig{
		HasHeader: true,
		Delimiter: ',',
	}
	
	for _, option := range options {
		option(config)
	}
	
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	reader.Comma = config.Delimiter
	
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}
	
	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}
	
	var columns []string
	var dataStart int
	
	if config.HasHeader {
		columns = records[0]
		dataStart = 1
	} else {
		columns = make([]string, len(records[0]))
		for i := range columns {
			columns[i] = fmt.Sprintf("col_%d", i)
		}
		dataStart = 0
	}
	
	df := NewDataFrame(columns)
	
	for i := dataStart; i < len(records); i++ {
		row := make([]interface{}, len(records[i]))
		for j, val := range records[i] {
			row[j] = inferType(val)
		}
		df.AddRow(row)
	}
	
	return df, nil
}

func (df *DataFrame) ToCSV(filename string, options ...CSVOption) error {
	config := &CSVConfig{
		HasHeader: true,
		Delimiter: ',',
	}
	
	for _, option := range options {
		option(config)
	}
	
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	writer.Comma = config.Delimiter
	defer writer.Flush()
	
	if config.HasHeader {
		if err := writer.Write(df.columns); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}
	
	for _, row := range df.data {
		stringRow := make([]string, len(row))
		for i, val := range row {
			stringRow[i] = fmt.Sprintf("%v", val)
		}
		if err := writer.Write(stringRow); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}
	
	return nil
}

type CSVConfig struct {
	HasHeader bool
	Delimiter rune
}

type CSVOption func(*CSVConfig)

func WithHeader(hasHeader bool) CSVOption {
	return func(c *CSVConfig) {
		c.HasHeader = hasHeader
	}
}

func WithDelimiter(delimiter rune) CSVOption {
	return func(c *CSVConfig) {
		c.Delimiter = delimiter
	}
}

func inferType(value string) interface{} {
	value = strings.TrimSpace(value)
	
	if value == "" {
		return nil
	}
	
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}
	
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}
	
	if boolVal, err := strconv.ParseBool(value); err == nil {
		return boolVal
	}
	
	return value
}