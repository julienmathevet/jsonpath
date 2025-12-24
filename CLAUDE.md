# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a specific test
go test -v -run TestParse ./...

# Build the CLI tool
go build -o jsonpath ./cmd/jsonpath

# Use the CLI tool
./jsonpath "$.store.book[0].title" data.json
cat data.json | ./jsonpath "$.store"
```

## Architecture

This is a Go library implementing JSONPath syntax for querying JSON data structures.

### Core Design Pattern

The parser uses a **chain of nodes** pattern where each JSONPath component becomes a node that processes data and passes results to the next node:

1. `Parse(path)` returns an `Applicator` interface
2. `Applicator.Apply(data)` traverses the node chain to filter JSON data
3. Each node type implements the `node` interface with `Apply()` and `SetNext()`

### Node Types (parse.go)

| Node | Purpose |
|------|---------|
| `RootNode` | Starting point, delegates to next node |
| `MapSelection` | Accesses map keys (e.g., `.store`) |
| `ArraySelection` | Accesses array indices (e.g., `[0]`, `[-1]`) |
| `WildCardSelection` | Returns all values (`[*]`) |
| `WildCardKeySelection` | Returns all keys (`[@]`) |
| `WildCardFilterSelection` | Filter expressions (`[?(@.price > 10)]`) |
| `DescentSelection` | Recursive descent (`..`) |

### Path Normalization

The `normalize()` function converts dot-notation to bracket-notation internally:
- `$.store.book[0]` becomes `$["store"]["book"][0]`

### Supported Operators

- `$` - Root element
- `@` - Current node in filter
- `*` - Wildcard for values
- `..` - Recursive descent
- `[n]` - Array index (supports negative)
- `[start:end]` - Array slice
- `[?(...)]` - Filter expressions with `<`, `<=`, `==`, `!=`, `>=`, `>`, `=~`, `!~`
- `||` - OR conditions in filters
