# JsonPath

A fast golang implementation of JSONPath syntax.
Follows the majority of rules from http://goessner.net/articles/JsonPath/
with some minor differences.

**Go Version Required**: 1.18+

## Get Started

```bash
go get github.com/julienmathevet/jsonpath
```

Example code:

```go
import (
    "encoding/json"
    "github.com/julienmathevet/jsonpath"
)

var json_data interface{}
json.Unmarshal([]byte(data), &json_data)

// Parse once, apply multiple times for best performance
filter, err := jsonpath.Parse("$.store.book[?(@.price < 10)]")
if err != nil {
    // handle error
}
result, err := filter.Apply(json_data)
```

## Performance

This library is optimized for performance with:
- Pre-compiled regex patterns
- Cached path parsing in filter operations
- Direct type comparisons (no reflection-based evaluation)
- Minimal memory allocations

Benchmark results show filter operations are **10x faster** than naive implementations,
with **95% fewer memory allocations**. See [BENCHMARK.md](BENCHMARK.md) for details.

## Operators

Referenced from github.com/jayway/JsonPath

| Operator | Supported | Description |
| ---- | :---: | ---------- |
| `$` | Y | The root element to query. This starts all path expressions. |
| `@` | Y | The current node being processed by a filter predicate. |
| `*` | Y | Wildcard. Available anywhere a name or numeric are required. |
| `..` | Y | Deep scan. Available anywhere a name is required. |
| `.<name>` | Y | Dot-notated child |
| `['<name>' (, '<name>')]` | X | Bracket-notated child or children |
| `[<number> (, <number>)]` | Y | Array index or indexes |
| `[start:end]` | Y | Array slice operator |
| `[?(<expression>)]` | Y | Filter expression. Expression must evaluate to a boolean value. |

### Filter Operators

| Operator | Description |
| -------- | ----------- |
| `<` | Less than |
| `<=` | Less than or equal |
| `==` | Equal |
| `!=` | Not equal |
| `>=` | Greater than or equal |
| `>` | Greater than |
| `=~` | Regex match |
| `!~` | Regex not match |
| `\|\|` | OR condition |

## Examples

Given this example data:

```json
{
    "store": {
        "book": [
            {
                "category": "reference",
                "author": "Nigel Rees",
                "title": "Sayings of the Century",
                "price": 8.95
            },
            {
                "category": "fiction",
                "author": "Evelyn Waugh",
                "title": "Sword of Honour",
                "price": 12.99
            },
            {
                "category": "fiction",
                "author": "Herman Melville",
                "title": "Moby Dick",
                "isbn": "0-553-21311-3",
                "price": 8.99
            },
            {
                "category": "fiction",
                "author": "J. R. R. Tolkien",
                "title": "The Lord of the Rings",
                "isbn": "0-395-19395-8",
                "price": 22.99
            }
        ],
        "bicycle": {
            "color": "red",
            "price": 19.95
        }
    },
    "expensive": 10
}
```

### Example JSONPath Queries

| JSONPath | Result |
| :--------- | :------- |
| `$.expensive` | `10` |
| `$.store.book[0].price` | `8.95` |
| `$.store.book[-1].isbn` | `"0-395-19395-8"` |
| `$.store.book[0,1].price` | `[8.95, 12.99]` |
| `$.store.book[0:2].price` | `[8.95, 12.99, 8.99]` |
| `$.store.book[?(@.isbn)].price` | `[8.99, 22.99]` |
| `$.store.book[?(@.price > 10)].title` | `["Sword of Honour", "The Lord of the Rings"]` |
| `$.store.book[?(@.price < $.expensive)].price` | `[8.95, 8.99]` |
| `$.store.book[:].price` | `[8.95, 12.99, 8.99, 22.99]` |
| `$..author` | `["Nigel Rees", "Evelyn Waugh", "Herman Melville", "J. R. R. Tolkien"]` |
| `$.store.book[?(@.author =~ 'J.*')]` | Books by authors starting with "J" |
| `$.store.book[?(@.category == 'fiction' \|\| @.price < 10)]` | Fiction books or books under $10 |

## License

MIT
