# Benchmark Results

## Environment
- **Platform**: darwin/arm64
- **CPU**: Apple M4 Pro
- **Go Version**: 1.25.4 (sections below), 1.27.0 (json/v2 migration section)

## Performance Comparison

### Before Optimization
```
BenchmarkParse-14                    	 6397029	       173.5 ns/op	     232 B/op	       9 allocs/op
BenchmarkNormalize-14                	11939199	        97.08 ns/op	      96 B/op	       4 allocs/op
BenchmarkApplySimple-14              	61606910	        19.28 ns/op	       0 B/op	       0 allocs/op
BenchmarkApplyArrayIndex-14          	61453633	        20.80 ns/op	       0 B/op	       0 allocs/op
BenchmarkApplyWildcard-14            	10126938	       106.7 ns/op	     136 B/op	       4 allocs/op
BenchmarkApplyDescent-14             	 1000000	      1120 ns/op	    1080 B/op	      26 allocs/op
BenchmarkApplyFilterNumeric-14       	   50280	     24415 ns/op	   59278 B/op	     639 allocs/op
BenchmarkApplyFilterString-14        	   49288	     23204 ns/op	   57462 B/op	     576 allocs/op
BenchmarkApplyFilterRegex-14         	   56739	     22068 ns/op	   55778 B/op	     593 allocs/op
BenchmarkCmpAny-14                   	  824310	      1480 ns/op	    3460 B/op	      52 allocs/op
BenchmarkCmpWildcard-14              	 1000000	      1002 ns/op	    2587 B/op	      39 allocs/op
BenchmarkApplyFilterLargeArray-14    	    2178	    556614 ns/op	 1416283 B/op	   14086 allocs/op
```

### After Optimization
```
BenchmarkParse-14                    	221324848	         5.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkParseSimplePath-14          	220441628	         5.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkParseCached-14              	222016924	         5.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkParseSimplePathUncached-14  	 21002336	        51.9 ns/op	     112 B/op	       4 allocs/op
BenchmarkNormalize-14                	 27166735	        44.8 ns/op	      48 B/op	       1 allocs/op
BenchmarkApplySimple-14              	 57844566	        19.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkApplyArrayIndex-14          	 58357478	        19.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkApplyWildcard-14            	 11511849	       106.1 ns/op	     136 B/op	       4 allocs/op
BenchmarkApplyDescent-14             	  1000000	      1118 ns/op	    1080 B/op	      26 allocs/op
BenchmarkApplyFilterNumeric-14       	   556282	      2230 ns/op	    1758 B/op	      31 allocs/op
BenchmarkApplyFilterString-14        	   404982	      2934 ns/op	    1847 B/op	      32 allocs/op
BenchmarkApplyFilterRegex-14         	   405853	      2845 ns/op	    1867 B/op	      37 allocs/op
BenchmarkCmpAny-14                   	 34649576	        35.8 ns/op	       2 B/op	       1 allocs/op
BenchmarkCmpWildcard-14              	  8077364	       146.1 ns/op	       5 B/op	       1 allocs/op
BenchmarkApplyFilterLargeArray-14    	    21962	     55165 ns/op	   44497 B/op	     708 allocs/op
```

## Summary

| Benchmark | Before | After | Speedup | Alloc Reduction |
|-----------|--------|-------|---------|-----------------|
| Parse (cached) | 173.5 ns | 5.4 ns | **32x** | 100% fewer |
| Parse (simple path, uncached) | 173.5 ns | 51.9 ns | **3.3x** | 56% fewer |
| Normalize | 97.08 ns | 44.8 ns | **2.2x** | 75% fewer |
| ApplyFilterNumeric | 24,415 ns | 2,230 ns | **11x** | 95% fewer |
| ApplyFilterString | 23,204 ns | 2,934 ns | **7.9x** | 94% fewer |
| ApplyFilterRegex | 22,068 ns | 2,845 ns | **7.8x** | 94% fewer |
| ApplyFilterLargeArray | 556,614 ns | 55,165 ns | **10x** | 95% fewer |
| CmpAny | 1,480 ns | 35.8 ns | **41x** | 98% fewer |
| CmpWildcard | 1,002 ns | 146.1 ns | **6.9x** | 97% fewer |

## Optimizations Applied

1. **Parse cache** - Cache parsed paths to avoid re-parsing (32x faster for repeated paths)
2. **Fast path for simple dot-notation** - Optimized parsing for `$.foo.bar` patterns (3.3x faster)
3. **Pre-compiled regexes** - Moved regex compilation to package initialization
4. **Replaced `types.Eval`** - Direct type-based comparisons instead of Go compiler evaluation
5. **Path caching in filters** - Cache parsed sub-paths in filter operations
6. **Regex caching** - Cache compiled wildcard patterns
7. **strings.Builder** - Reduced string allocations in normalize function

## encoding/json v2 Migration (Go 1.27)

The library core never calls `encoding/json` - it consumes already-decoded
`map[string]any` / `[]any` values. So the `Parse`/`Apply`
benchmarks above are **unaffected** by the v1 -> v2 switch. Re-running the
unchanged suite back-to-back on the same binary produced a +2.18% geomean
swing, which is the machine noise floor here; any apparent movement in those
numbers from this migration is drift, not signal.

What the switch does move is the decode step every caller performs first,
covered by `BenchmarkUnmarshal` and `BenchmarkUnmarshalThenApply`
(`benchtime=300ms`, `count=6`, benchstat):

| Benchmark | encoding/json | encoding/json/v2 | Delta |
|-----------|---------------|------------------|-------|
| Unmarshal (time) | 3.911 µs | 1.936 µs | **-50.5%** |
| Unmarshal (bytes) | 3.338 KiB | 2.939 KiB | -12.0% |
| Unmarshal (allocs) | 86 | 68 | -20.9% |
| UnmarshalThenApply (time) | 5.861 µs | 3.810 µs | **-35.0%** |
| UnmarshalThenApply (bytes) | 4.177 KiB | 3.763 KiB | -9.9% |
| UnmarshalThenApply (allocs) | 113 | 95 | -15.9% |

Standalone comparison on a larger payload (100-element array, same shape as
`BenchmarkApplyFilterLargeArray`): `Unmarshal` 49.46 µs -> 23.74 µs (-52%),
47.77 KiB -> 42.26 KiB, 1115 -> 812 allocs.

CLI indented output (`json.MarshalIndent` -> `json.Marshal` with
`jsontext.WithIndent` + `json.Deterministic(true)`), which is byte-identical
to the v1 output:

| Payload | v1 | v2 | Delta |
|---------|----|----|-------|
| store doc | 2.360 µs / 1.431 KiB / 9 allocs | 2.024 µs / 985 B / 8 allocs | -14% time, -33% bytes |
| 100-element array | 26.40 µs / 13.33 KiB / 6 allocs | 23.80 µs / 9.325 KiB / 5 allocs | -10% time, -30% bytes |

## Benchmark Loop Style

The suite uses `b.Loop()` (Go 1.24+) rather than `for i := 0; i < b.N; i++`.
`b.Loop()` keeps loop values alive, so it can defeat dead-code elimination that
a `b.N` loop would allow, and it excludes setup from the measured time (making
the previous `b.ResetTimer()` calls redundant). Converting the existing
benchmarks moved every number by at most 6%, i.e. within this machine's noise
floor, so the tables above remain representative rather than having been
inflated by eliminated work.

## Test Coverage

- **Before**: 87.4%
- **After**: 92.1%
