package jsonpath

import (
	"encoding/json/v2"
	"testing"
)

var benchmarkData map[string]any

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

// benchmarkJSON is the same document as benchmarkData, kept as raw bytes so the
// decode step callers must perform is measured too. The library itself never
// touches encoding/json - it consumes already-decoded values - so these two
// benchmarks are what a json/v2 change actually moves.
var benchmarkJSON = []byte(`{
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
}`)

func BenchmarkUnmarshal(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var v map[string]any
		if err := json.Unmarshal(benchmarkJSON, &v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalThenApply(b *testing.B) {
	a, err := Parse("$.store.book[?(@.price < 10)]")
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		var v map[string]any
		if err := json.Unmarshal(benchmarkJSON, &v); err != nil {
			b.Fatal(err)
		}
		if _, err := a.Apply(v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = Parse("$.store.book[0].title")
	}
}

func BenchmarkParseSimplePath(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = Parse("$.store.bicycle.color")
	}
}

func BenchmarkParseCached(b *testing.B) {
	// Pre-warm the cache
	_, _ = Parse("$.store.book.title")
	b.ReportAllocs()
	for b.Loop() {
		_, _ = Parse("$.store.book.title")
	}
}

func BenchmarkParseSimplePathUncached(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		// Use unique paths to avoid cache hits
		_ = parseSimpleDotPath("$.store.bicycle.color")
	}
}

func BenchmarkParseNoCache(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = ParseNoCache("$.store.bicycle.color")
	}
}

func BenchmarkNormalize(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = normalize("$.store.book[0].title")
	}
}

func BenchmarkApplySimple(b *testing.B) {
	a, _ := Parse("$.store.bicycle.color")
	b.ReportAllocs()
	for b.Loop() {
		_, _ = a.Apply(benchmarkData)
	}
}

func BenchmarkApplyArrayIndex(b *testing.B) {
	a, _ := Parse("$.store.book[0].title")
	b.ReportAllocs()
	for b.Loop() {
		_, _ = a.Apply(benchmarkData)
	}
}

func BenchmarkApplyWildcard(b *testing.B) {
	a, _ := Parse("$.store.book[*].author")
	b.ReportAllocs()
	for b.Loop() {
		_, _ = a.Apply(benchmarkData)
	}
}

func BenchmarkApplyDescent(b *testing.B) {
	a, _ := Parse("$..author")
	b.ReportAllocs()
	for b.Loop() {
		_, _ = a.Apply(benchmarkData)
	}
}

func BenchmarkApplyFilterNumeric(b *testing.B) {
	a, _ := Parse("$.store.book[?(@.price < 10)]")
	b.ReportAllocs()
	for b.Loop() {
		_, _ = a.Apply(benchmarkData)
	}
}

func BenchmarkApplyFilterString(b *testing.B) {
	a, _ := Parse("$.store.book[?(@.category == fiction)]")
	b.ReportAllocs()
	for b.Loop() {
		_, _ = a.Apply(benchmarkData)
	}
}

func BenchmarkApplyFilterRegex(b *testing.B) {
	a, _ := Parse("$.store.book[?(@.author =~ 'J.*')]")
	b.ReportAllocs()
	for b.Loop() {
		_, _ = a.Apply(benchmarkData)
	}
}

func BenchmarkCmpAny(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = cmp_any(8.95, "10", "<")
	}
}

func BenchmarkCmpWildcard(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = cmp_wildcard("J. R. R. Tolkien", "J.*", "=~")
	}
}

// Benchmark with larger dataset to show filter performance impact
func BenchmarkApplyFilterLargeArray(b *testing.B) {
	// Create a large array of items
	largeData := make(map[string]any)
	items := make([]any, 100)
	for i := range 100 {
		items[i] = map[string]any{
			"id":    i,
			"price": float64(i * 10),
			"name":  "item",
		}
	}
	largeData["items"] = items

	a, _ := Parse("$.items[?(@.price < 500)]")
	b.ReportAllocs()
	for b.Loop() {
		_, _ = a.Apply(largeData)
	}
}
