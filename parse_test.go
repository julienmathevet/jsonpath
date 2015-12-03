package jsonpath

import "testing"

func TestNormalize(t *testing.T) {
	var ev string
	var err error
	testcases := []struct {
		t                 string // Test Value
		e                 string // Expected Value
		err               error  // Expected Error
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
	}
	for i, test := range testcases {
		// First let's run the normal test.
		ev, err = normalize(test.t)
		if ev != test.e || err != test.err {
			t.Errorf("[%03d] Standard test:  Normalize(\"%v\") = \"%v\", %v; want %v, %v", i, test.t, ev, err, test.e, test.err)
		}
		// Test that a normalized version is a noop.
		ev, err = normalize(test.e)
		if ev != test.e || err != test.err {
			t.Errorf("[%03d] NOOP test:      Normalize(\"%v\") = \"%v\", %v; want %v, %v", i, test.e, ev, err, test.e, test.err)
		}
		if !test.DontTestImplicate {
			// Test that an implicate version works as well.
			ev, err = normalize(test.t[2:])
			if ev != test.e || err != test.err {
				t.Errorf("[%03d] Implicate test: Normalize(\"%v\") = \"%v\", %v; want %v, %v", i, test.t[2:], ev, err, test.e, test.err)
			}
		}

	}

}
