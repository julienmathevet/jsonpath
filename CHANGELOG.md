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
- **Performance**: Filter operations are now up to 10x faster
- **Performance**: `cmp_any` comparison is 40x faster (replaced `go/types.Eval` with direct comparisons)
- **Performance**: `cmp_wildcard` regex matching is 6.7x faster (cached regex compilation)
- **Performance**: `normalize` function is 2x faster (using `strings.Builder`)
- **Performance**: Memory allocations reduced by 95% in filter operations

### Fixed
- Regex compilation no longer happens on every filter call
- Sub-path parsing is now cached to avoid redundant parsing

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
