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
BenchmarkParse-14                    	 9383493	       123.6 ns/op	     184 B/op	       6 allocs/op
BenchmarkNormalize-14                	26136625	        44.61 ns/op	      48 B/op	       1 allocs/op
BenchmarkApplySimple-14              	58037954	        20.32 ns/op	       0 B/op	       0 allocs/op
BenchmarkApplyArrayIndex-14          	55579248	        21.05 ns/op	       0 B/op	       0 allocs/op
BenchmarkApplyWildcard-14            	10941014	       111.6 ns/op	     136 B/op	       4 allocs/op
BenchmarkApplyDescent-14             	 1000000	      1148 ns/op	    1080 B/op	      26 allocs/op
BenchmarkApplyFilterNumeric-14       	  536479	      2332 ns/op	    1759 B/op	      31 allocs/op
BenchmarkApplyFilterString-14        	  393814	      3031 ns/op	    1847 B/op	      32 allocs/op
BenchmarkApplyFilterRegex-14         	  398064	      3001 ns/op	    1865 B/op	      37 allocs/op
BenchmarkCmpAny-14                   	33637471	        36.23 ns/op	       2 B/op	       1 allocs/op
BenchmarkCmpWildcard-14              	 8067866	       149.5 ns/op	       5 B/op	       1 allocs/op
BenchmarkApplyFilterLargeArray-14    	   20858	     56579 ns/op	   44489 B/op	     708 allocs/op
```

## Summary

| Benchmark | Before | After | Speedup | Alloc Reduction |
|-----------|--------|-------|---------|-----------------|
| Parse | 173.5 ns | 123.6 ns | **1.4x** | 33% fewer |
| Normalize | 97.08 ns | 44.61 ns | **2.2x** | 75% fewer |
| ApplyFilterNumeric | 24,415 ns | 2,332 ns | **10.5x** | 95% fewer |
| ApplyFilterString | 23,204 ns | 3,031 ns | **7.7x** | 94% fewer |
| ApplyFilterRegex | 22,068 ns | 3,001 ns | **7.4x** | 94% fewer |
| ApplyFilterLargeArray | 556,614 ns | 56,579 ns | **9.8x** | 95% fewer |
| CmpAny | 1,480 ns | 36.23 ns | **40.9x** | 98% fewer |
| CmpWildcard | 1,002 ns | 149.5 ns | **6.7x** | 97% fewer |

## Optimizations Applied

1. **Pre-compiled regexes** - Moved regex compilation to package initialization
2. **Replaced `types.Eval`** - Direct type-based comparisons instead of Go compiler evaluation
3. **Path caching** - Cache parsed sub-paths in filter operations
4. **Regex caching** - Cache compiled wildcard patterns
5. **strings.Builder** - Reduced string allocations in normalize function

## Test Coverage

- **Before**: 87.4%
- **After**: 92.1%
