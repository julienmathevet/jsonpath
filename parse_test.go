package jsonpath

import "testing"

func TestNormalize(t *testing.T) {
	var ev string
	var ee error
	testcases := []struct {
		tv string // Test Value
		ev string // Expected Value
		ee error  // Expected Error
	}{
		{"$.store.book[0].title", `$["store"]["book"][0]["title"]`, nil},
		/*
			{"$.store.book[*].author", `$["store"]["book"][*]["author"]`, nil},
			{"$..author", `$[..]["author"]`, nil},
			{"$.store.*", `$["store"][*]`, nil},
			{"$.store.book[?(@.length()-1)].title", `$["store"]["book"][?(@.length()-1)["title"]`, nil},
		*/
	}
	for i, test := range testcases {
		// First let's run the normal test.
		ev, ee = normalize(test.tv)
		if ev != test.ev || ee != test.ee {
			t.Errorf("[%03d] Standard test:  Normalize(\"%v\") = \"%v\", %v; want %v, %v", i, test.tv, ev, ee, test.ev, test.ee)
		}
		// Test that a normalized version is a noop.
		ev, ee = normalize(test.ev)
		if ev != test.ev || ee != test.ee {
			t.Errorf("[%03d] NOOP test:      Normalize(\"%v\") = \"%v\", %v; want %v, %v", i, test.ev, ev, ee, test.ev, test.ee)
		}
		// Test that an implicate version works as well.
		ev, ee = normalize(test.tv[2:])
		if ev != test.ev || ee != test.ee {
			t.Errorf("[%03d] Implicate test: Normalize(\"%v\") = \"%v\", %v; want %v, %v", i, test.tv[2:], ev, ee, test.ev, test.ee)
		}

	}

}
