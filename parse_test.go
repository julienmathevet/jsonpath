package jsonpath

import "testing"

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
		return mv.Key == n.Key
	default:
		return false
	}
}
func isSameMapSelectionNode(n *MapSelection, m node) bool {
	switch mv := m.(type) {
	case *MapSelection:
		return mv.Key == n.Key
	default:
		return false
	}
}
func isSameRootNodeNode(n *RootNode, m node) bool {
	switch m.(type) {
	case *RootNode:
		return true
	default:
		return false
	}
}
func isSameWildcardSelectionNode(n *WildCardSelection, m node) bool {
	switch m.(type) {
	case *WildCardSelection:
		return true
	default:
		return false
	}
}
func isSameDescentSelectionNode(n *DescentSelection, m node) bool {
	switch m.(type) {
	case *DescentSelection:
		return true
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
		{t: `[?(@.lenght())]`, n: nil, err: NotSupportedError},
		{t: `[(@.foo)]`, n: nil, err: NotSupportedError},
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
