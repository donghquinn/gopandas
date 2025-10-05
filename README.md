# Go-Pandas

A Go library for data manipulation and analysis, inspired by Python's pandas library. Provides DataFrame and Series data structures with essential data processing capabilities, all implemented without external dependencies.

## Features

- **DataFrame and Series** - Core data structures for handling structured data
- **CSV Support** - Read and write CSV files with automatic type inference
- **Excel Support** - Read Excel files (.xlsx) without external dependencies
- **Data Operations** - Filter, select, sort, and group data
- **Statistical Functions** - Calculate sum, mean, count, and more
- **Zero Dependencies** - Pure Go implementation

## Installation

```bash
go get github.com/donghyun/go-pandas
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    gopandas "github.com/donghyun/go-pandas"
)

func main() {
    // Read CSV file
    df, err := gopandas.ReadCSV("data.csv")
    if err != nil {
        log.Fatal(err)
    }

    // Display basic info
    rows, cols := df.Shape()
    fmt.Printf("Shape: (%d, %d)\n", rows, cols)
    fmt.Printf("Columns: %v\n", df.Columns())
    
    // Show first 5 rows
    fmt.Print(df.Head(5))
}
```

## Core Data Structures

### DataFrame

A 2-dimensional labeled data structure with columns of potentially different types.

```go
// Create a new DataFrame
df := gopandas.NewDataFrame([]string{"name", "age", "city"})

// Add rows
df.AddRow([]interface{}{"Alice", 25, "New York"})
df.AddRow([]interface{}{"Bob", 30, "London"})

// Get shape
rows, cols := df.Shape()

// Get column names
columns := df.Columns()

// Display first n rows
head := df.Head(3)
```

### Series

A 1-dimensional labeled array capable of holding any data type.

```go
// Create a new Series
data := []interface{}{1, 2, 3, 4, 5}
series := gopandas.NewSeries("numbers", data)

// Statistical operations
sum, _ := series.Sum()      // 15.0
mean, _ := series.Mean()    // 3.0
count := series.Count()     // 5
```

## File I/O

### CSV Operations

```go
// Read CSV with default options (header=true, delimiter=',')
df, err := gopandas.ReadCSV("data.csv")

// Read CSV with custom options
df, err := gopandas.ReadCSV("data.csv", 
    gopandas.WithHeader(false),
    gopandas.WithDelimiter(';'))

// Write to CSV
err = df.ToCSV("output.csv")

// Write CSV with custom options
err = df.ToCSV("output.csv",
    gopandas.WithHeader(true),
    gopandas.WithDelimiter(','))
```

### Excel Operations

```go
// Read Excel file (first sheet)
df, err := gopandas.ReadExcel("data.xlsx")

// Read specific sheet
df, err := gopandas.ReadExcel("data.xlsx", "Sheet2")
```

## Data Manipulation

### Filtering

```go
// Filter rows based on condition
filtered := df.Filter(func(row []interface{}) bool {
    age := row[1].(int)
    return age >= 30
})
```

### Column Selection

```go
// Select specific columns
subset, err := df.Select("name", "age")
```

### Sorting

```go
// Sort by column (ascending)
sorted, err := df.Sort("age", true)

// Sort by column (descending)
sorted, err := df.Sort("salary", false)
```

### Grouping

```go
// Group by column
groups, err := df.GroupBy("department")

// Iterate through groups
for key, group := range groups {
    fmt.Printf("Group %v:\n", key)
    fmt.Print(group)
}
```

### Column Operations

```go
// Get a column as Series
ageColumn, err := df.GetColumn("age")

// Calculate statistics
avgAge, err := ageColumn.Mean()
totalAge, err := ageColumn.Sum()
count := ageColumn.Count()
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    gopandas "github.com/donghyun/go-pandas"
)

func main() {
    // Create sample data
    df := gopandas.NewDataFrame([]string{"name", "age", "department", "salary"})
    df.AddRow([]interface{}{"Alice", 25, "Engineering", 70000})
    df.AddRow([]interface{}{"Bob", 30, "Sales", 50000})
    df.AddRow([]interface{}{"Charlie", 35, "Engineering", 80000})
    df.AddRow([]interface{}{"Diana", 28, "Marketing", 55000})

    // Display basic information
    rows, cols := df.Shape()
    fmt.Printf("Dataset shape: (%d, %d)\n", rows, cols)
    fmt.Print(df)

    // Filter engineering employees
    engineers := df.Filter(func(row []interface{}) bool {
        return row[2].(string) == "Engineering"
    })
    fmt.Println("\nEngineering employees:")
    fmt.Print(engineers)

    // Calculate average salary
    salaryColumn, _ := df.GetColumn("salary")
    avgSalary, _ := salaryColumn.Mean()
    fmt.Printf("\nAverage salary: $%.2f\n", avgSalary)

    // Group by department
    groups, _ := df.GroupBy("department")
    fmt.Println("\nEmployees by department:")
    for dept, group := range groups {
        rows, _ := group.Shape()
        fmt.Printf("%s: %d employees\n", dept, rows)
    }

    // Save to CSV
    err := df.ToCSV("employees.csv")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Data saved to employees.csv")
}
```

## API Reference

### DataFrame Methods

- `NewDataFrame(columns []string) *DataFrame` - Create new DataFrame
- `Shape() (int, int)` - Get number of rows and columns
- `Columns() []string` - Get column names
- `Head(n int) *DataFrame` - Get first n rows
- `AddRow(row []interface{}) error` - Add a new row
- `GetColumn(name string) (*Series, error)` - Get column as Series
- `Filter(predicate func([]interface{}) bool) *DataFrame` - Filter rows
- `Select(columns ...string) (*DataFrame, error)` - Select columns
- `Sort(column string, ascending bool) (*DataFrame, error)` - Sort by column
- `GroupBy(column string) (map[interface{}]*DataFrame, error)` - Group by column
- `ToCSV(filename string, options ...CSVOption) error` - Write to CSV

### Series Methods

- `NewSeries(name string, data []interface{}) *Series` - Create new Series
- `Sum() (interface{}, error)` - Calculate sum
- `Mean() (float64, error)` - Calculate mean
- `Count() int` - Count non-null values

### File I/O Functions

- `ReadCSV(filename string, options ...CSVOption) (*DataFrame, error)` - Read CSV
- `ReadExcel(filename string, sheetName ...string) (*DataFrame, error)` - Read Excel

### CSV Options

- `WithHeader(hasHeader bool)` - Set header option
- `WithDelimiter(delimiter rune)` - Set delimiter

## Testing

Run tests:

```bash
go test
```

Run example:

```bash
cd example
go run main.go
```

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.