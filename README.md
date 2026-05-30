# gopandas

[![Test](https://github.com/donghquinn/gopandas/actions/workflows/test.yml/badge.svg)](https://github.com/donghquinn/gopandas/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/donghquinn/gopandas/branch/main/graph/badge.svg)](https://codecov.io/gh/donghquinn/gopandas)

A Go library for data manipulation and analysis, inspired by Python's pandas. Provides `DataFrame` and `Series` data structures with data processing, statistics, and file I/O — all implemented without external dependencies.

## Features

- **DataFrame and Series** — core data structures for structured data
- **File I/O** — read/write CSV, Excel (.xlsx/.xls), and JSON
- **Data manipulation** — filter, select, sort, group, merge, concat
- **Missing value handling** — DropNA, FillNA, IsNull, NotNull
- **Statistics** — sum, mean, std, variance, median, min, max, percentiles
- **Column operations** — rename, drop, add, apply transformations
- **Row access** — Head, Tail, Iloc (slice/position-based)
- **Zero external dependencies** — pure Go standard library

## Installation

```bash
go get github.com/donghquinn/gopandas
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    gopandas "github.com/donghquinn/gopandas"
)

func main() {
    df, err := gopandas.ReadCSV("data.csv")
    if err != nil {
        log.Fatal(err)
    }

    rows, cols := df.Shape()
    fmt.Printf("Shape: (%d, %d)\n", rows, cols)
    fmt.Printf("Columns: %v\n", df.Columns())
    fmt.Print(df.Head(5))
}
```

## Core Data Structures

### DataFrame

A 2-dimensional labeled data structure with columns of potentially different types.

```go
df := gopandas.NewDataFrame([]string{"name", "age", "city"})

df.AddRow([]interface{}{"Alice", 25, "New York"})
df.AddRow([]interface{}{"Bob", 30, "London"})
df.AddRow([]interface{}{"Charlie", 35, "Paris"})

rows, cols := df.Shape()       // (3, 3)
columns := df.Columns()        // ["name", "age", "city"]

fmt.Print(df.Head(2))          // first 2 rows
fmt.Print(df.Tail(2))          // last 2 rows
fmt.Print(df)                  // all rows as a table
```

### Series

A 1-dimensional labeled array capable of holding any data type.

```go
s := gopandas.NewSeries("scores", []interface{}{85, 92, 78, 95, 88})

sum, _    := s.Sum()     // 438.0
mean, _   := s.Mean()    // 87.6
count     := s.Count()   // 5
max, _    := s.Max()     // 95
min, _    := s.Min()     // 78
median, _ := s.Median()  // 88.0
std, _    := s.Std()     // sample standard deviation
v, _      := s.Var()     // sample variance
```

## File I/O

### CSV

```go
// Read (auto-detects types: int, float64, bool, string)
df, err := gopandas.ReadCSV("data.csv")

// Read with options
df, err := gopandas.ReadCSV("data.tsv",
    gopandas.WithHeader(true),
    gopandas.WithDelimiter('\t'))

// Write
err = df.ToCSV("output.csv")
err = df.ToCSV("output.tsv", gopandas.WithDelimiter('\t'))
```

### Excel

```go
// Read first sheet
df, err := gopandas.ReadExcel("data.xlsx")

// Read a specific sheet
df, err := gopandas.ReadExcel("data.xlsx", "Sheet2")

// Both .xlsx and .xls formats are supported
df, err := gopandas.ReadExcel("legacy.xls")
```

### JSON

JSON files must be an array of objects. Keys from all records are used as columns.

```go
// Read
df, err := gopandas.ReadJSON("data.json")

// Write
err = df.ToJSON("output.json")
```

Example JSON format:
```json
[
  {"name": "Alice", "age": 25, "city": "New York"},
  {"name": "Bob",   "age": 30, "city": "London"}
]
```

## Data Manipulation

### Filter

```go
// Keep rows where age >= 30
adults := df.Filter(func(row []interface{}) bool {
    return row[1].(int) >= 30
})
```

### Select

```go
// Select specific columns
subset, err := df.Select("name", "salary")
```

### Sort

```go
sorted, err := df.Sort("salary", false) // descending
sorted, err := df.Sort("name", true)    // ascending
```

### GroupBy

```go
groups, err := df.GroupBy("department")
for dept, group := range groups {
    rows, _ := group.Shape()
    fmt.Printf("%s: %d rows\n", dept, rows)
}
```

### Apply (row-wise)

```go
// Double every salary value
updated := df.Apply(func(row []interface{}) []interface{} {
    newRow := make([]interface{}, len(row))
    copy(newRow, row)
    newRow[3] = row[3].(int) * 2
    return newRow
})
```

## Column Operations

```go
// Get a column as Series
ages, err := df.GetColumn("age")

// Rename columns
df2 := df.Rename(map[string]string{"age": "years", "name": "full_name"})

// Drop columns
df3, err := df.Drop("city", "zip")

// Add or overwrite a column
err = df.SetColumn("senior", []interface{}{false, true, true})

// Get column type map
types := df.Dtypes()
// map["name":"string", "age":"int", ...]
```

## Row Access

```go
// First / last n rows
top3    := df.Head(3)
bottom3 := df.Tail(3)

// Slice by integer range [start, end) — supports negative indices
middle, err := df.Iloc(2, 5)
last2, err  := df.Iloc(-2, -1)

// Specific row positions
picked, err := df.IlocRows(0, 3, 7)

// Row index values
idx := df.Index()
```

## Combining DataFrames

### Concat (vertical stack)

All DataFrames must have the same columns in the same order.

```go
combined, err := gopandas.Concat(df1, df2, df3)
```

### Merge (join)

Supports `"inner"`, `"left"`, `"right"`, and `"outer"` joins on a common column.

```go
// Inner join
result, err := employees.Merge(departments, "dept_id", "inner")

// Left join — all rows from left, matched rows from right
result, err := orders.Merge(customers, "customer_id", "left")

// Outer join — all rows from both sides
result, err := df1.Merge(df2, "id", "outer")
```

When both DataFrames have a non-key column with the same name, the right-side column is suffixed with `_right`.

## Missing Values

```go
// DataFrame
clean   := df.DropNA()           // drop rows containing any nil
filled  := df.FillNA(0)          // replace nil with 0
hasNull := df.HasNA()            // true if any nil exists

// Series
nullMask := s.IsNull()           // []bool — true where nil
notNull  := s.NotNull()          // []bool — true where not nil
clean    := s.DropNA()           // new Series with nils removed
filled   := s.FillNA(0)         // new Series with nils replaced
```

## Statistics

### Series

```go
s := gopandas.NewSeries("data", []interface{}{2, 4, 4, 4, 5, 5, 7, 9})

max, _    := s.Max()             // 9
min, _    := s.Min()             // 2
median, _ := s.Median()          // 4.5
std, _    := s.Std()             // sample std (ddof=1)
v, _      := s.Var()             // sample variance (ddof=1)

unique    := s.Unique()          // []interface{}{2, 4, 5, 7, 9}
n         := s.NUnique()         // 5
counts    := s.ValueCounts()     // map[interface{}]int{4:3, 5:2, ...}

// Apply a function element-wise
doubled := s.Apply(func(v interface{}) interface{} {
    return v.(int) * 2
})

// Accessor helpers
name   := s.Name()    // "data"
values := s.Values()  // []interface{}{...}
```

### DataFrame.Describe

Returns a summary DataFrame with 8 statistics for every numeric column.

```go
desc := df.Describe()
fmt.Print(desc)
// stat           age            salary
// -----------------------------------------------
// count          4              4
// mean           29.5           63750
// std            4.43           12500.0
// min            25             50000
// 25%            27.25          53750
// 50%            29.5           65000
// 75%            31.75          72500
// max            35             80000
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    gopandas "github.com/donghquinn/gopandas"
)

func main() {
    // Build a DataFrame manually
    df := gopandas.NewDataFrame([]string{"name", "dept", "salary"})
    df.AddRow([]interface{}{"Alice",   "Engineering", 80000})
    df.AddRow([]interface{}{"Bob",     "Sales",       50000})
    df.AddRow([]interface{}{"Charlie", "Engineering", 90000})
    df.AddRow([]interface{}{"Diana",   "Sales",       55000})
    df.AddRow([]interface{}{"Eve",     "Marketing",   nil})

    // Basic info
    rows, cols := df.Shape()
    fmt.Printf("Shape: (%d, %d)\n", rows, cols)

    // Drop rows with missing data
    clean := df.DropNA()

    // Statistics summary
    fmt.Print(clean.Describe())

    // Filter and sort
    highEarners, _ := clean.Filter(func(row []interface{}) bool {
        return row[2].(int) >= 70000
    }), nil
    sorted, _ := highEarners.Sort("salary", false)
    fmt.Print(sorted)

    // Group by department and compute mean salary
    groups, _ := clean.GroupBy("dept")
    for dept, g := range groups {
        col, _ := g.GetColumn("salary")
        avg, _ := col.Mean()
        fmt.Printf("%s avg salary: $%.0f\n", dept, avg)
    }

    // Merge with a departments DataFrame
    depts := gopandas.NewDataFrame([]string{"dept", "head_count"})
    depts.AddRow([]interface{}{"Engineering", 120})
    depts.AddRow([]interface{}{"Sales",       80})

    merged, err := clean.Merge(depts, "dept", "left")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Print(merged)

    // Save results
    merged.ToCSV("result.csv")
    merged.ToJSON("result.json")
}
```

## API Reference

### DataFrame

| Method | Description |
|--------|-------------|
| `NewDataFrame(columns []string) *DataFrame` | Create a new empty DataFrame |
| `AddRow(row []interface{}) error` | Append a row |
| `Shape() (int, int)` | Row and column count |
| `Columns() []string` | Column names |
| `Index() []interface{}` | Row index values |
| `Dtypes() map[string]string` | Inferred Go type per column |
| `Head(n int) *DataFrame` | First n rows |
| `Tail(n int) *DataFrame` | Last n rows |
| `Iloc(start, end int) (*DataFrame, error)` | Rows by position range [start, end) |
| `IlocRows(indices ...int) (*DataFrame, error)` | Rows at specific positions |
| `Filter(fn func([]interface{}) bool) *DataFrame` | Keep rows matching predicate |
| `Select(columns ...string) (*DataFrame, error)` | Select columns by name |
| `Drop(columns ...string) (*DataFrame, error)` | Remove columns |
| `Rename(mapping map[string]string) *DataFrame` | Rename columns |
| `SetColumn(name string, values []interface{}) error` | Add or replace a column |
| `GetColumn(name string) (*Series, error)` | Column as Series |
| `Sort(column string, ascending bool) (*DataFrame, error)` | Sort rows by column |
| `GroupBy(column string) (map[interface{}]*DataFrame, error)` | Group rows by column value |
| `Apply(fn func([]interface{}) []interface{}) *DataFrame` | Apply function to each row |
| `Describe() *DataFrame` | Summary statistics for numeric columns |
| `DropNA() *DataFrame` | Remove rows containing nil |
| `FillNA(value interface{}) *DataFrame` | Replace nil values |
| `HasNA() bool` | True if any nil exists |
| `String() string` | Tabular string representation |
| `ToCSV(filename string, options ...CSVOption) error` | Write to CSV |
| `ToJSON(filename string) error` | Write to JSON |

### Package-level functions

| Function | Description |
|----------|-------------|
| `NewDataFrame(columns []string) *DataFrame` | Create DataFrame |
| `NewSeries(name string, data []interface{}) *Series` | Create Series |
| `Concat(dfs ...*DataFrame) (*DataFrame, error)` | Stack DataFrames vertically |
| `ReadCSV(filename string, options ...CSVOption) (*DataFrame, error)` | Read CSV |
| `ReadExcel(filename string, sheetName ...string) (*DataFrame, error)` | Read Excel |
| `ReadJSON(filename string) (*DataFrame, error)` | Read JSON array-of-objects |

### DataFrame.Merge

```go
func (df *DataFrame) Merge(right *DataFrame, on string, how string) (*DataFrame, error)
```

`how` must be one of `"inner"`, `"left"`, `"right"`, or `"outer"`.

### Series

| Method | Description |
|--------|-------------|
| `Name() string` | Series name |
| `Values() []interface{}` | Raw data slice |
| `Count() int` | Non-nil value count |
| `Sum() (interface{}, error)` | Sum of numeric values |
| `Mean() (float64, error)` | Arithmetic mean |
| `Median() (float64, error)` | Median |
| `Std() (float64, error)` | Sample standard deviation (ddof=1) |
| `Var() (float64, error)` | Sample variance (ddof=1) |
| `Max() (interface{}, error)` | Maximum value |
| `Min() (interface{}, error)` | Minimum value |
| `Unique() []interface{}` | Distinct values, first-seen order |
| `NUnique() int` | Count of distinct non-nil values |
| `ValueCounts() map[interface{}]int` | Frequency of each value |
| `Apply(fn func(interface{}) interface{}) *Series` | Element-wise transform |
| `IsNull() []bool` | Nil mask |
| `NotNull() []bool` | Non-nil mask |
| `DropNA() *Series` | Remove nil values |
| `FillNA(value interface{}) *Series` | Replace nil values |

### CSV Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithHeader(bool)` | `true` | First row is a header |
| `WithDelimiter(rune)` | `','` | Field separator |

## Testing

```bash
go test ./...
```

## License

MIT License — see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome. Please open an issue or submit a pull request.
