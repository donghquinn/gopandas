package gopandas

import (
	"fmt"
	"testing"
)

func benchDataFrame(rows int) *DataFrame {
	df := NewDataFrame([]string{"id", "name", "score", "active"})
	for i := 0; i < rows; i++ {
		df.AddRow([]interface{}{i, fmt.Sprintf("user_%d", i%100), float64(i) * 1.5, i%2 == 0})
	}
	return df
}

func BenchmarkString(b *testing.B) {
	df := benchDataFrame(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = df.String()
	}
}

func BenchmarkDescribe(b *testing.B) {
	df := benchDataFrame(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = df.Describe()
	}
}

func BenchmarkFilter(b *testing.B) {
	df := benchDataFrame(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = df.Filter(func(row []interface{}) bool { return row[0].(int)%2 == 0 })
	}
}

func BenchmarkSelect(b *testing.B) {
	df := benchDataFrame(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = df.Select("id", "score")
	}
}

func BenchmarkGetColumn(b *testing.B) {
	df := benchDataFrame(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = df.GetColumn("score")
	}
}

func BenchmarkSeriesMean(b *testing.B) {
	df := benchDataFrame(10000)
	s, _ := df.GetColumn("score")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.Mean()
	}
}

func BenchmarkSort(b *testing.B) {
	df := benchDataFrame(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = df.Sort("name", true)
	}
}

func BenchmarkGroupBy(b *testing.B) {
	df := benchDataFrame(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = df.GroupBy("name")
	}
}

func BenchmarkMerge(b *testing.B) {
	left := benchDataFrame(5000)
	right := NewDataFrame([]string{"name", "dept"})
	for i := 0; i < 100; i++ {
		right.AddRow([]interface{}{fmt.Sprintf("user_%d", i), fmt.Sprintf("dept_%d", i%10)})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = left.Merge(right, "name", "inner")
	}
}

func BenchmarkInferType(b *testing.B) {
	inputs := []string{"12345", "3.14159", "true", "hello world", "user_42", ""}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range inputs {
			_ = inferType(s)
		}
	}
}
