# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added
- Comprehensive test suite with 92% code coverage
- Benchmark suite for performance testing
- Regex filter operators `=~` and `!~` for pattern matching in filters
- OR conditions support in filters using `||`

### Changed
- **Performance**: Parse with cache is 32x faster (5.4 ns vs 173.5 ns)
- **Performance**: Simple dot-notation paths (`$.foo.bar`) are 3.3x faster with dedicated fast path
- **Performance**: Filter operations are now up to 10x faster
- **Performance**: `cmp_any` comparison is 40x faster (replaced `go/types.Eval` with direct comparisons)
- **Performance**: `cmp_wildcard` regex matching is 6.7x faster (cached regex compilation)
- **Performance**: `normalize` function is 2x faster (using `strings.Builder`)
- **Performance**: Memory allocations reduced by 95% in filter operations

### Fixed
- Regex compilation no longer happens on every filter call
- Sub-path parsing is now cached to avoid redundant parsing
- Filter flow logic: `=~` and `!~` operators no longer incorrectly fall through to `cmp_any`
- Condition checks: fixed `||` to `&&` for proper error/nil handling in wildcard selections
- `cmp_wildcard` type switch: now correctly matches pattern from `obj2` parameter
- `minNotNeg1` function: now correctly finds minimum instead of returning last value when first arg is -1
- Potential panic in `normalize`: added bounds check before accessing string index
- Thread-safety: `WildCardFilterSelection.pathCache` now protected with mutex
- Thread-safety: `getCachedPath` now returns errors and uses proper locking
- Malformed paths: `Parse` now returns `ErrSyntax` for invalid paths (e.g., unclosed brackets) instead of silently returning a partial result

## [1.0.0] - Previous

### Added
- Initial JSONPath implementation
- Support for root element `$`
- Support for current node `@` in filters
- Wildcard `*` selection
- Deep scan `..` operator
- Dot-notation and bracket-notation
- Array index and slice operators
- Filter expressions with comparison operators (`<`, `<=`, `==`, `!=`, `>=`, `>`)
