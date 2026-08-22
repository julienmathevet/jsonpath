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
- **Breaking**: Minimum Go version is now 1.27 (was 1.25), required by `encoding/json/v2`
- Migrated from `encoding/json` to `encoding/json/v2` in the CLI and tests. The library
  itself is JSON-decoder agnostic: it operates on already-decoded
  `map[string]interface{}` / `[]interface{}` values, so callers may use either package.
- **Performance**: decoding with `encoding/json/v2` is ~2x faster than v1 (`Unmarshal`
  into `map[string]interface{}`: 4.01µs -> 2.00µs on the readme's store document,
  49.5µs -> 23.7µs on a 100-element array), with ~20-27% fewer allocations
- **Performance**: indented marshaling in the CLI is ~10-14% faster with ~30% less
  memory than `json.MarshalIndent`
- **Performance**: Parse with cache is 32x faster (5.4 ns vs 173.5 ns)
- **Performance**: Simple dot-notation paths (`$.foo.bar`) are 3.3x faster with dedicated fast path
- **Performance**: Filter operations are now up to 10x faster
- **Performance**: `cmp_any` comparison is 40x faster (replaced `go/types.Eval` with direct comparisons)
- **Performance**: `cmp_wildcard` regex matching is 6.7x faster (cached regex compilation)
- **Performance**: `normalize` function is 2x faster (using `strings.Builder`)
- **Performance**: Memory allocations reduced by 95% in filter operations

### Fixed
- CLI: `encoding/json/v2` does not sort map keys by default, so `json.Deterministic(true)`
  is passed explicitly to keep the CLI's output byte-identical to the v1 build
- Replaced deprecated `io/ioutil` with `io`/`os` in the CLI
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
