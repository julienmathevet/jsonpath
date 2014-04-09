package jsonpath

import (
	"encoding/json"
	"testing"
)

// Simple test case from json_path website:
var jsonBooksBlob = []byte(`
{ "store": {
    "book": [ 
      { "category": "reference",
        "author": "Nigel Rees",
        "title": "Sayings of the Century",
        "price": 8.95
      },
      { "category": "fiction",
        "author": "Evelyn Waugh",
        "title": "Sword of Honour",
        "price": 12.99
      },
      { "category": "fiction",
        "author": "Herman Melville",
        "title": "Moby Dick",
        "isbn": "0-553-21311-3",
        "price": 8.99
      },
      { "category": "fiction",
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
  }
}
`)
var bookstore interface{}
var bs Object

func init() {
	// set up the fixtures
	json.Unmarshal(jsonBooksBlob, &bookstore)
	bs = Object(bookstore)
}

func helperTestStringPath(t *testing.T, o Object, path, eResponse, desc string) {
	v, err := Value(path, o)
	if err != nil {
		t.Errorf("We got an error trying to %s [ %s]", desc, err)
	}
	s, found := v.(string)
	if !found {
		t.Error("Did not get a string back for path %s", path)
	}
	if s != eResponse {
		t.Errorf("Did not get the expected string '%s' got '%s' instead ", eResponse, s)
	}
}
func TestGetValue(t *testing.T) {

	helperTestStringPath(t, bs, `$['store']['book'][0]['title']`, "Sayings of the Century", "title of the first book")
	helperTestStringPath(t, bs, `$['store']['book'][1]['title']`, "Sword of Honour", "title of the second book")
	helperTestStringPath(t, bs, `$['store']['bicycle']['color']`, "red", "color of the bicycle")
}

/*
func TestParsePath(t *testing.T) {
	p, err := ParsePath(`$`)
	if err != nil {
		t.Error("Did not expect to recieve an error: ", err)
	}
	if len(p) != 0 {
		t.Error("For a root element expected path to be empty")
	}
	p, err = ParsePath(`$['store']`)
	if err != nil {
		t.Error("Did not expect to recieve an error: ", err)
	}
	if len(p) != 1 {
		t.Error("Expected path to contain one element")
	}
	e := p[0]
	if e.Type != MapType {
		t.Error("Expected first element to be a MapKey")
	}
	if e.MapKey != "store" {
		t.Error("Expected key to be `store` go ", e.MapKey)
	}
}
*/
func hlperTestNormalizePath(t *testing.T, path, expectedPath string) {
	p, err := normalize(path)
	if err != nil {
		t.Errorf("Did not expect error normalizing `%s` error: `%v`", path, err)
	}
	if p != expectedPath {
		t.Errorf("Normalizing `%s` should be `%s` got `%s`", path, expectedPath, p)
	}

}
func TestNormalizePath(t *testing.T) {
	_, err := normalize(`store`)
	if err == nil {
		t.Errorf("Expected normalize to return error.")
	}
	fix := map[string]string{
		`  $  `:                               `$`,
		`$`:                                   `$`,
		`  $.book`:                            `$['book']`,
		`$.store  `:                           `$['store']`,
		`$['book']`:                           `$['book']`,
		`$['store']`:                          `$['store']`,
		`$['book'][0]`:                        `$['book'][0]`,
		`$.store[0]`:                          `$['store'][0]`,
		`$.store..price`:                      `$['store'][..]['price']`,
		`$.store[0]..price`:                   `$['store'][0][..]['price']`,
		`$.store..*`:                          `$['store'][..][*]`,
		`$.store[..][*]`:                      `$['store'][..][*]`,
		`$['store'][..][*]`:                   `$['store'][..][*]`,
		`$..book[?(@.isbn)]`:                  `$[..]['book'][?(@.isbn)]`,
		`$..book[(@.length-1)]`:               `$[..]['book'][(@.length-1)]`,
		`$..book[(@.bar[@.foo[1]].length-1)]`: `$[..]['book'][(@.bar[@.foo[1]].length-1)]`,
		`$..book['bar[]']`:                    `$[..]['book']['bar[]']`,
		`$.book['bar()']`:                     `$['book']['bar()']`,
		`$.['store','price']..*`:              `$['store','price'][..][*]`,
		`$['book'][0:-1]`:                     `$['book'][0:-1]`,
		`$['book'][:2]`:                       `$['book'][:2]`,
		`$.book.[0:-1:2]`:                     `$['book'][0:-1:2]`,
		`$.book.0`:                            `$['book'][0]`,
		`$.book'\''.0`:                        `$['book'\'''][0]`,
		`$.book.10`:                           `$['book'][10]`,
		`$.book.0day`:                         `$['book']['0day']`,
		`$.book.10day4Nov2014`:                `$['book']['10day4Nov2014']`,
		`$.book.10day.4Nov.'2014'`:            `$['book']['10day']['4Nov']['2014']`,
		`$..`:       `$[..]`,
		`$..*`:      `$[..][*]`,
		`$.'foo\'s`: `$['foo\'s']`,
	}
	for k, v := range fix {
		hlperTestNormalizePath(t, k, v)
	}
}
