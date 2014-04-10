package jsonpath

import (
	"testing"
)

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
	_, err = normalize(`$['store[[0]']`)
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
