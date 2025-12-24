package jsonpath

import (
	"encoding/json"
	"testing"
)

var benchmarkData map[string]interface{}

func init() {
	err := json.Unmarshal([]byte(`{
		"store": {
			"book": [
				{"category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95},
				{"category": "fiction", "author": "Evelyn Waugh", "title": "Sword of Honor", "price": 12.99},
				{"category": "fiction", "author": "Herman Melville", "title": "Moby Dick", "isbn": "0-553-21311-3", "price": 8.99},
				{"category": "fiction", "author": "J. R. R. Tolkien", "title": "The Lord of the Rings", "isbn": "0-395-19395-8", "price": 22.99}
			],
			"bicycle": {"color": "red", "price": 19.95}
		},
		"expensive": 10
	}`), &benchmarkData)
	if err != nil {
		panic(err)
	}
}

func BenchmarkParse(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = Parse("$.store.book[0].title")
	}
}

func BenchmarkParseSimplePath(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = Parse("$.store.bicycle.color")
	}
}

func BenchmarkParseCached(b *testing.B) {
	// Pre-warm the cache
	Parse("$.store.book.title")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse("$.store.book.title")
	}
}

func BenchmarkParseSimplePathUncached(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Use unique paths to avoid cache hits
		_ = parseSimpleDotPath("$.store.bicycle.color")
	}
}

func BenchmarkNormalize(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = normalize("$.store.book[0].title")
	}
}

func BenchmarkApplySimple(b *testing.B) {
	a, _ := Parse("$.store.bicycle.color")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.Apply(benchmarkData)
	}
}

func BenchmarkApplyArrayIndex(b *testing.B) {
	a, _ := Parse("$.store.book[0].title")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.Apply(benchmarkData)
	}
}

func BenchmarkApplyWildcard(b *testing.B) {
	a, _ := Parse("$.store.book[*].author")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.Apply(benchmarkData)
	}
}

func BenchmarkApplyDescent(b *testing.B) {
	a, _ := Parse("$..author")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.Apply(benchmarkData)
	}
}

func BenchmarkApplyFilterNumeric(b *testing.B) {
	a, _ := Parse("$.store.book[?(@.price < 10)]")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.Apply(benchmarkData)
	}
}

func BenchmarkApplyFilterString(b *testing.B) {
	a, _ := Parse("$.store.book[?(@.category == fiction)]")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.Apply(benchmarkData)
	}
}

func BenchmarkApplyFilterRegex(b *testing.B) {
	a, _ := Parse("$.store.book[?(@.author =~ 'J.*')]")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.Apply(benchmarkData)
	}
}

func BenchmarkCmpAny(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = cmp_any(8.95, "10", "<")
	}
}

func BenchmarkCmpWildcard(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = cmp_wildcard("J. R. R. Tolkien", "J.*", "=~")
	}
}

// Benchmark with larger dataset to show filter performance impact
func BenchmarkApplyFilterLargeArray(b *testing.B) {
	// Create a large array of items
	largeData := make(map[string]interface{})
	items := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		items[i] = map[string]interface{}{
			"id":    i,
			"price": float64(i * 10),
			"name":  "item",
		}
	}
	largeData["items"] = items

	a, _ := Parse("$.items[?(@.price < 500)]")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.Apply(largeData)
	}
}
