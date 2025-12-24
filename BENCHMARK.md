# Benchmark Results

## Environment
- **Platform**: darwin/arm64
- **CPU**: Apple M4 Pro
- **Go Version**: 1.25.4

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

## Test Coverage

- **Before**: 87.4%
- **After**: 92.1%
