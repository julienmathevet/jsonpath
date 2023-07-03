package jsonpath

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestParseWildcardCondition(t *testing.T) {
	w := WildCardFilterSelection{Key: "@.metadata.project_name =~ 'A.*'"}
	r := map[string]interface{}{
		"metadata": map[string]interface{}{
			"project_name": "AProject",
		},
	}
	res, err := w.filter(r)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if res == nil {
		t.Errorf("filter() = nil; want true")
	}
}

func TestParseWildcardConditionDoesNotMatch(t *testing.T) {
	w := WildCardFilterSelection{Key: "@.metadata.project_name =~ 'A.*'"}
	r := map[string]interface{}{
		"metadata": map[string]interface{}{
			"project_name": "BProject",
		},
	}
	res, err := w.filter(r)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if res != nil {
		t.Errorf("filter() = %v; want nil", res)
	}
}

func TestParseWildcardConditionDifferentOf(t *testing.T) {
	w := WildCardFilterSelection{Key: "@.metadata.project_name !~ 'A.*'"}
	r := map[string]interface{}{
		"metadata": map[string]interface{}{
			"project_name": "BProject",
		},
	}
	res, err := w.filter(r)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if res == nil {
		t.Errorf("filter() = nil; want true")
	}
}

func TestParseWildcardConditionWithDotes(t *testing.T) {
	w := WildCardFilterSelection{Key: "@.context.releaseVersion != ''"}

	r := map[string]interface{}{
		"context": map[string]interface{}{
			"releaseVersion": "1.0.0",
		},
	}
	res, err := w.filter(r)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if res == nil {
		t.Errorf("filter() = nil; want true")
	}
}

func TestParseWildcardConditionWithBackslash(t *testing.T) {
	w := WildCardFilterSelection{Key: "@.context.releaseVersion != ''"}

	r := map[string]interface{}{
		"context": map[string]interface{}{
			"releaseVersion": "FCL-11 \\ \\ QAMLess",
		},
	}
	res, err := w.filter(r)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if res == nil {
		t.Errorf("filter() = nil; want true")
	}
}

func TestParseWildcardConditionWithDoubleQuotes(t *testing.T) {
	w := WildCardFilterSelection{Key: "@.context.releaseVersion != ''"}

	r := map[string]interface{}{
		"context": map[string]interface{}{
			"releaseVersion": "FCL-11 \" \" QAMLess",
		},
	}
	res, err := w.filter(r)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if res == nil {
		t.Errorf("filter() = nil; want true")
	}
}

func TestGetConditionsFromKeySimple(t *testing.T) {
	w := WildCardFilterSelection{Key: "@.metadata.project_name != 'DEV-QA-WEB-BROWSER'"}
	conditions, err := w.GetConditionsFromKey()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if len(conditions) != 1 {
		t.Errorf("getConditionsFromKey() = %v; want 1", len(conditions))
	}
	expected := `@.metadata.project_name != 'DEV-QA-WEB-BROWSER'`
	// verify expected matches another string
	if conditions[0] != expected {
		t.Errorf("getConditionsFromKey() = %v; want %v", conditions[0], expected)
	}
}

func TestGetConditionsFromKeyOrCondition(t *testing.T) {
	w := WildCardFilterSelection{Key: "@.metadata.project_name != 'DEV-QA-WEB-BROWSER' || @.metadata.project_name == 'DEV-QA-WEB-BROWSER2'"}
	conditions, err := w.GetConditionsFromKey()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if len(conditions) != 2 {
		t.Errorf("getConditionsFromKey() = %v; want 1", len(conditions))
	}
	expected1 := `@.metadata.project_name != 'DEV-QA-WEB-BROWSER'`
	expected2 := `@.metadata.project_name == 'DEV-QA-WEB-BROWSER2'`
	// verify expected matches another string
	if conditions[0] != expected1 {
		t.Errorf("getConditionsFromKey() = %v; want %v", conditions[0], expected1)
	}
	if conditions[1] != expected2 {
		t.Errorf("getConditionsFromKey() = %v; want %v", conditions[1], expected2)
	}
}

func TestNormalize(t *testing.T) {
	var ev string
	testcases := []struct {
		t                 string // Test Value
		e                 string // Expected Value
		DontTestImplicate bool
	}{
		{t: "$.store.book[0].title", e: `$["store"]["book"][0]["title"]`},
		{t: "$.store.book[*].author", e: `$["store"]["book"][*]["author"]`},
		{t: "$.", e: `$`},
		{t: ".", e: "$", DontTestImplicate: true},
		{t: "$..author", e: `$[..]["author"]`, DontTestImplicate: true},
		{t: "..author", e: `$[..]["author"]`, DontTestImplicate: true},
		{t: "$.store.*", e: `$["store"][*]`},
		{t: "$.store.book[?(@.length()-1)].title", e: `$["store"]["book"][?(@.length()-1)]["title"]`},
		{t: "$.store.book[1:10:2].title", e: `$["store"]["book"][1:10:2]["title"]`},
		{t: `$.store["category.sub"]`, e: `$["store"]["category.sub"]`},
		{t: `$.store..["category.sub"]`, e: `$["store"][..]["category.sub"]`},
		{t: `$.store..`, e: `$["store"][..]`},
		{t: `$.store.`, e: `$["store"]`},
		{t: `.store.`, e: `$["store"]`, DontTestImplicate: true},
		{t: `..store.`, e: `$[..]["store"]`, DontTestImplicate: true},
	}
	for i, test := range testcases {
		// First let's run the normal test.
		if ev = normalize(test.t); ev != test.e {
			t.Errorf("[%03d] Standard test:  Normalize(\"%v\") = \"%v\"; want \"%v\"", i, test.t, ev, test.e)
		}
		// Test that a normalized version is a noop.
		if ev = normalize(test.e); ev != test.e {
			t.Errorf("[%03d] NOOP test:      Normalize(\"%v\") = \"%v\"; want \"%v\"", i, test.e, ev, test.e)
		}
		if !test.DontTestImplicate {
			// Test that an implicate version works as well.
			if ev = normalize(test.t[2:]); ev != test.e {
				t.Errorf("[%03d] Implicate test: Normalize(\"%v\") = \"%v\"; want \"%v\"", i, test.t[2:], ev, test.e)
			}
		}
	}
}

func isSameArraySelectionNode(n *ArraySelection, m node) bool {
	switch mv := m.(type) {
	case *ArraySelection:
		if mv.Key != n.Key {
			return false
		}
		return isSameNode(n.NextNode, mv.NextNode)
	default:
		return false
	}
}
func isSameMapSelectionNode(n *MapSelection, m node) bool {
	switch mv := m.(type) {
	case *MapSelection:
		if mv.Key != n.Key {
			return false
		}
		return isSameNode(n.NextNode, mv.NextNode)
	default:
		return false
	}
}
func isSameRootNodeNode(n *RootNode, m node) bool {
	switch mv := m.(type) {
	case *RootNode:
		return isSameNode(n.NextNode, mv.NextNode)
	default:
		return false
	}
}
func isSameWildcardSelectionNode(n *WildCardSelection, m node) bool {
	switch mv := m.(type) {
	case *WildCardSelection:
		return isSameNode(n.NextNode, mv.NextNode)
	default:
		return false
	}
}
func isSameWildcardFilterSelectionNode(n *WildCardFilterSelection, m node) bool {
	switch mv := m.(type) {
	case *WildCardFilterSelection:
		return isSameNode(n.NextNode, mv.NextNode)
	default:
		return false
	}
}
func isSameDescentSelectionNode(n *DescentSelection, m node) bool {
	switch mv := m.(type) {
	case *DescentSelection:
		return isSameNode(n.NextNode, mv.NextNode)
	default:
		return false
	}
}

func isSameNode(n, m node) bool {
	if n == nil && m == nil {
		return true
	}
	switch nv := n.(type) {
	case *MapSelection:
		return isSameMapSelectionNode(nv, m)
	case *ArraySelection:
		return isSameArraySelectionNode(nv, m)
	case *RootNode:
		return isSameRootNodeNode(nv, m)
	case *DescentSelection:
		return isSameDescentSelectionNode(nv, m)
	case *WildCardSelection:
		return isSameWildcardSelectionNode(nv, m)
	case *WildCardFilterSelection:
		return isSameWildcardFilterSelectionNode(nv, m)
	default:
		return false
	}
}

func TestGetNode(t *testing.T) {
	testcases := []struct {
		t   string
		n   node
		s   string
		err error
	}{
		{t: `["store"]`, n: &MapSelection{Key: "store"}},
		{t: `[10]`, n: &ArraySelection{Key: 10}},
		{t: `[*]`, n: &WildCardSelection{}},
		{t: `[..]`, n: &DescentSelection{}},
		{t: `[?(@.lenght())]`, n: &WildCardFilterSelection{Key: "@.lenght()"}},
		{t: `[(@.foo)]`, n: &WildCardFilterSelection{Key: "@.foo"}},
		{t: `[0:10:2]`, n: nil, err: SyntaxError},
	}
	for i, test := range testcases {
		n, s, err := getNode(test.t)
		b := isSameNode(n, test.n)
		if err != test.err || s != test.s || !b {
			t.Errorf(`[%03d] getNode("%v") = %T, "%v","%v"; want %T, "%v", "%v"`, i, test.t, n, s, err, test.n, test.s, test.err)
		}
	}
}

func collapseNodes(n ...node) node {
	rt := &RootNode{}
	for i := range n {
		rt.SetNext(n[i])
	}
	return rt
}

func TestParse(t *testing.T) {
	var books = make(map[string]interface{})
	err := json.Unmarshal(
		[]byte(`{ "store" :
  { "book" :
    [ { "category"     : "reference"
      , "category.sub" : "quotes"
      , "author"       : "Nigel Rees"
      , "title"        : "Saying of the Century"
      , "price"        : 8.95
      }
    , { "category" : "fiction"
      , "author"   : "Evelyn Waugh"
      , "title"    : "Sword of Honor"
      , "price"    : 12.99
      }
    , { "category" : "fiction"
      , "author"   : "Herman Melville"
      , "title"    : "Moby Dick"
      , "isbn"     : "0-553-21311-3"
      , "price"    : 8.99
      }
    , { "category" : "fiction"
      , "author"   : "J. R. R. Tolkien"
      , "title"    : "The Lord of the Rings"
      , "isbn"     : "0-395-19395-8"
      , "price"    : 22.99
	  , "discount": true
      }
    ]
  , "bicycle" :
    { "color" : "red"
    , "price" : 19.95
    }
  }
}`),
		&books,
	)
	if err != nil {
		t.Fatal("Test Cast Parse error: ", err)
		return
	}

	testcases := []struct {
		t        string
		expected interface{}
		perr     error
		err      error
	}{
		{
			t:        "$..author",
			expected: []interface{}{"Nigel Rees", "Evelyn Waugh", "Herman Melville", "J. R. R. Tolkien"},
		},
		{
			t: "store.bicycle",
			expected: map[string]interface{}{
				"color": "red",
				"price": 19.95,
			},
		},
		{
			t: "store.bicycle.*",
			expected: []interface{}{
				"red",
				19.95,
			},
		},
		{
			t: "store.bicycle.@",
			expected: []interface{}{
				"color",
				"price",
			},
		},
		{
			t: "store.book[0]",
			expected: map[string]interface{}{
				"category":     "reference",
				"category.sub": "quotes",
				"author":       "Nigel Rees",
				"title":        "Saying of the Century",
				"price":        8.95,
			},
		},
		{
			t: "store.book[0]*",
			expected: []interface{}{
				"Nigel Rees",
				"reference",
				"quotes",
				8.95,
				"Saying of the Century",
			},
		},
		{
			t: "store.book..isbn",
			expected: []interface{}{
				"0-553-21311-3",
				"0-395-19395-8",
			},
		},
		{
			t: "store..isbn",
			expected: []interface{}{
				"0-553-21311-3",
				"0-395-19395-8",
			},
		},
		{
			t: "store..[\"category.sub\"]",
			expected: []interface{}{
				"quotes",
			},
		},
		{
			t:        "..author..",
			expected: []interface{}{"Nigel Rees", "Evelyn Waugh", "Herman Melville", "J. R. R. Tolkien"},
		},
		{
			t: "store.book[?(@.price < 10)]",
			expected: []interface{}{map[string]interface{}{
				"category":     "reference",
				"category.sub": "quotes",
				"author":       "Nigel Rees",
				"title":        "Saying of the Century",
				"price":        8.95,
			}, map[string]interface{}{
				"category": "fiction",
				"author":   "Herman Melville",
				"title":    "Moby Dick",
				"isbn":     "0-553-21311-3",
				"price":    8.99,
			},
			},
		},
		{
			t: `store.book[?(@.category == reference)]`,
			expected: []interface{}{map[string]interface{}{
				"category":     "reference",
				"category.sub": "quotes",
				"author":       "Nigel Rees",
				"title":        "Saying of the Century",
				"price":        8.95,
			},
			},
		},
		{
			t: `store.book[?(@.discount)]`,
			expected: []interface{}{map[string]interface{}{
				"category": "fiction",
				"author":   "J. R. R. Tolkien",
				"title":    "The Lord of the Rings",
				"isbn":     "0-395-19395-8",
				"price":    22.99,
				"discount": true,
			},
			},
		},
		{
			t: `store.book[?(@.author == 'Nigel Rees')]`,
			expected: []interface{}{map[string]interface{}{
				"category":     "reference",
				"category.sub": "quotes",
				"author":       "Nigel Rees",
				"title":        "Saying of the Century",
				"price":        8.95,
			},
			},
		},
	}
	for i, test := range testcases {
		a, err := Parse(test.t)
		if err != test.perr {
			t.Errorf(`[%03d] Parse("%v") = %T, %v ; expected err to be %v`, i, test.t, a, err, test.perr)
			continue
		}
		ev, err := a.Apply(books)
		if err != test.err {
			t.Errorf(`[%03d] %v a.Apply(books) = %T, %v ; expected err to be %v`, i, test.t, ev, err, test.err)
			continue
		}
		evs := fmt.Sprintf("“%v”", ev)
		tevs := fmt.Sprintf("“%v”", test.expected)
		if !reflect.DeepEqual(ev, test.expected) {
			t.Errorf(`[%03d] a.Apply(books) = %v, nil ; expected to be %v`, i, evs, tevs)
		}
	}
}
