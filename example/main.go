package main

import (
	"fmt"
	"log"
	"os"

	"github.com/donghquinn/gopandas"
)

func main() {
	fmt.Println("gopandas Example Usage")
	fmt.Println("=======================")

	createSampleData()
	demonstrateCSV()
	demonstrateDataFrame()
	demonstrateOperations()
}

func createSampleData() {
	csvData := `name,age,department,salary
Alice,25,Engineering,70000
Bob,30,Sales,50000
Charlie,35,Engineering,80000
Diana,28,Marketing,55000
Eve,32,Sales,60000
Frank,29,Engineering,75000`

	err := os.WriteFile("sample.csv", []byte(csvData), 0644)
	if err != nil {
		log.Fatal("Failed to create sample CSV:", err)
	}
	fmt.Println("✓ Sample CSV file created")
}

func demonstrateCSV() {
	fmt.Println("\n1. Reading CSV File")
	fmt.Println("-------------------")

	df, err := gopandas.ReadCSV("sample.csv")
	if err != nil {
		log.Fatal("Failed to read CSV:", err)
	}

	rows, cols := df.Shape()
	fmt.Printf("DataFrame shape: (%d, %d)\n", rows, cols)
	fmt.Printf("Columns: %v\n", df.Columns())
	fmt.Println("\nFirst 3 rows:")
	fmt.Print(df.Head(3))
}

func demonstrateDataFrame() {
	fmt.Println("\n2. DataFrame Operations")
	fmt.Println("-----------------------")

	df := gopandas.NewDataFrame([]string{"product", "price", "quantity"})
	df.AddRow([]interface{}{"Apple", 1.2, 100})
	df.AddRow([]interface{}{"Banana", 0.8, 150})
	df.AddRow([]interface{}{"Orange", 1.5, 80})

	fmt.Println("Product DataFrame:")
	fmt.Print(df)

	priceColumn, err := df.GetColumn("price")
	if err != nil {
		log.Fatal("Failed to get price column:", err)
	}

	avgPrice, err := priceColumn.Mean()
	if err != nil {
		log.Fatal("Failed to calculate average price:", err)
	}

	fmt.Printf("Average price: $%.2f\n", avgPrice)
}

func demonstrateOperations() {
	fmt.Println("\n3. Data Manipulation")
	fmt.Println("--------------------")

	df, err := gopandas.ReadCSV("sample.csv")
	if err != nil {
		log.Fatal("Failed to read CSV:", err)
	}

	fmt.Println("Original data:")
	fmt.Print(df.Head(3))

	engineeringOnly := df.Filter(func(row []interface{}) bool {
		return row[2].(string) == "Engineering"
	})

	fmt.Println("\nEngineering department only:")
	fmt.Print(engineeringOnly)

	nameAndSalary, err := df.Select("name", "salary")
	if err != nil {
		log.Fatal("Failed to select columns:", err)
	}

	fmt.Println("\nName and Salary only:")
	fmt.Print(nameAndSalary.Head(3))

	sortedBySalary, err := df.Sort("salary", false)
	if err != nil {
		log.Fatal("Failed to sort by salary:", err)
	}

	fmt.Println("\nSorted by salary (descending):")
	fmt.Print(sortedBySalary.Head(3))

	groups, err := df.GroupBy("department")
	if err != nil {
		log.Fatal("Failed to group by department:", err)
	}

	fmt.Println("\nGrouped by department:")
	for dept, group := range groups {
		rows, _ := group.Shape()
		fmt.Printf("\n%s (%d employees):\n", dept, rows)
		fmt.Print(group)
	}

	salaryColumn, err := df.GetColumn("salary")
	if err != nil {
		log.Fatal("Failed to get salary column:", err)
	}

	avgSalary, err := salaryColumn.Mean()
	if err != nil {
		log.Fatal("Failed to calculate average salary:", err)
	}

	fmt.Printf("\nAverage salary across all departments: $%.2f\n", avgSalary)

	os.Remove("sample.csv")
	fmt.Println("\n✓ Sample file cleaned up")
}
