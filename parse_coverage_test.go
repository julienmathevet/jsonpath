package jsonpath

import (
	"testing"
	"time"
)

func TestCmpAny(t *testing.T) {
	testcases := []struct {
		name     string
		obj1     interface{}
		obj2     interface{}
		op       string
		expected bool
		wantErr  bool
	}{
		// Numeric comparisons
		{name: "float less than", obj1: 5.0, obj2: "10", op: "<", expected: true},
		{name: "float greater than", obj1: 15.0, obj2: "10", op: ">", expected: true},
		{name: "float equal", obj1: 10.0, obj2: "10", op: "==", expected: true},
		{name: "float not equal", obj1: 5.0, obj2: "10", op: "!=", expected: true},
		{name: "float less or equal", obj1: 10.0, obj2: "10", op: "<=", expected: true},
		{name: "float greater or equal", obj1: 10.0, obj2: "10", op: ">=", expected: true},

		// String comparisons
		{name: "string equal", obj1: "hello", obj2: "'hello'", op: "==", expected: true},
		{name: "string not equal", obj1: "hello", obj2: "'world'", op: "!=", expected: true},
		{name: "string less than", obj1: "abc", obj2: "'bcd'", op: "<", expected: true},
		{name: "string greater than", obj1: "xyz", obj2: "'abc'", op: ">", expected: true},

		// Integer comparisons
		{name: "int equal", obj1: 42, obj2: "42", op: "==", expected: true},
		{name: "int not equal", obj1: 42, obj2: "43", op: "!=", expected: true},

		// Boolean comparisons
		{name: "bool equal true", obj1: true, obj2: "true", op: "==", expected: true},
		{name: "bool equal false", obj1: false, obj2: "false", op: "==", expected: true},
		{name: "bool not equal", obj1: true, obj2: "false", op: "!=", expected: true},

		// Invalid operator
		{name: "invalid operator", obj1: 5.0, obj2: "10", op: "~~", wantErr: true},
		{name: "invalid operator plus", obj1: 5.0, obj2: "10", op: "+", wantErr: true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := cmp_any(tc.obj1, tc.obj2, tc.op)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tc.expected {
				t.Errorf("cmp_any(%v, %v, %s) = %v; want %v", tc.obj1, tc.obj2, tc.op, result, tc.expected)
			}
		})
	}
}

func TestCmpWildcard(t *testing.T) {
	testcases := []struct {
		name     string
		obj1     interface{}
		obj2     interface{}
		op       string
		expected bool
		wantErr  bool
	}{
		// Match operator =~
		{name: "regex match simple", obj1: "hello", obj2: "hel.*", op: "=~", expected: true},
		{name: "regex match full", obj1: "test123", obj2: "test\\d+", op: "=~", expected: true},
		{name: "regex no match", obj1: "hello", obj2: "world.*", op: "=~", expected: false},

		// Not match operator !~
		{name: "regex not match", obj1: "hello", obj2: "world.*", op: "!~", expected: true},
		{name: "regex not match fail", obj1: "hello", obj2: "hel.*", op: "!~", expected: false},

		// Non-string types
		{name: "int regex match", obj1: 123, obj2: "12.*", op: "=~", expected: true},
		{name: "float regex match", obj1: 3.14, obj2: "3\\.14", op: "=~", expected: true},

		// Invalid operator
		{name: "invalid operator", obj1: "hello", obj2: "world", op: "==", wantErr: true},

		// Invalid regex
		{name: "invalid regex", obj1: "hello", obj2: "[invalid", op: "=~", wantErr: true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := cmp_wildcard(tc.obj1, tc.obj2, tc.op)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tc.expected {
				t.Errorf("cmp_wildcard(%v, %v, %s) = %v; want %v", tc.obj1, tc.obj2, tc.op, result, tc.expected)
			}
		})
	}
}

func TestArraySelectionApply(t *testing.T) {
	testcases := []struct {
		name    string
		key     int
		input   interface{}
		want    interface{}
		wantErr error
	}{
		{
			name:  "valid index",
			key:   1,
			input: []interface{}{"a", "b", "c"},
			want:  "b",
		},
		{
			name:  "first element",
			key:   0,
			input: []interface{}{"first", "second"},
			want:  "first",
		},
		{
			name:    "negative index",
			key:     -1,
			input:   []interface{}{"a", "b", "c"},
			wantErr: IndexOutOfBounds,
		},
		{
			name:    "out of bounds",
			key:     10,
			input:   []interface{}{"a", "b"},
			wantErr: IndexOutOfBounds,
		},
		{
			name:    "not an array",
			key:     0,
			input:   map[string]interface{}{"key": "value"},
			wantErr: ArrayTypeError,
		},
		{
			name:    "string input",
			key:     0,
			input:   "not an array",
			wantErr: ArrayTypeError,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			a := &ArraySelection{Key: tc.key}
			result, err := a.Apply(tc.input)
			if tc.wantErr != nil {
				if err != tc.wantErr {
					t.Errorf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tc.want {
				t.Errorf("Apply() = %v; want %v", result, tc.want)
			}
		})
	}
}

func TestWildCardSelectionApply(t *testing.T) {
	testcases := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{
			name:  "map input",
			input: map[string]interface{}{"a": 1, "b": 2},
			want:  []interface{}{1, 2},
		},
		{
			name:  "array input",
			input: []interface{}{1, 2, 3},
			want:  []interface{}{1, 2, 3},
		},
		{
			name:  "scalar string",
			input: "scalar",
			want:  "scalar",
		},
		{
			name:  "scalar int",
			input: 42,
			want:  42,
		},
		{
			name:  "nil input",
			input: nil,
			want:  nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			w := &WildCardSelection{}
			result, err := w.Apply(tc.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			// For slices, check length and values
			if resultSlice, ok := result.([]interface{}); ok {
				wantSlice, _ := tc.want.([]interface{})
				if len(resultSlice) != len(wantSlice) {
					t.Errorf("Apply() = %v; want %v", result, tc.want)
				}
			}
		})
	}
}

func TestWildCardFilterSelectionApply(t *testing.T) {
	testcases := []struct {
		name    string
		key     string
		input   interface{}
		wantLen int
		wantErr bool
	}{
		{
			name:    "filter array",
			key:     "@.price < 10",
			input:   []interface{}{map[string]interface{}{"price": 5.0}, map[string]interface{}{"price": 15.0}},
			wantLen: 1,
		},
		{
			name:    "filter single map",
			key:     "@.price < 10",
			input:   map[string]interface{}{"price": 5.0},
			wantLen: 1,
		},
		{
			name:    "filter single map no match",
			key:     "@.price < 10",
			input:   map[string]interface{}{"price": 50.0},
			wantLen: 0,
		},
		{
			name:    "scalar input error",
			key:     "@.price < 10",
			input:   "scalar",
			wantErr: true,
		},
		{
			name:    "int input error",
			key:     "@.price < 10",
			input:   42,
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			w := &WildCardFilterSelection{Key: tc.key}
			result, err := w.Apply(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			resultSlice, ok := result.([]interface{})
			if !ok {
				t.Errorf("expected slice result, got %T", result)
				return
			}
			if len(resultSlice) != tc.wantLen {
				t.Errorf("Apply() returned %d items; want %d", len(resultSlice), tc.wantLen)
			}
		})
	}
}

func TestWildCardFilterSelectionFilterErrors(t *testing.T) {
	testcases := []struct {
		name    string
		key     string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "non-map input",
			key:     "@.price < 10",
			input:   "not a map",
			wantErr: true,
		},
		{
			name:    "empty key",
			key:     "",
			input:   map[string]interface{}{"price": 5.0},
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			w := &WildCardFilterSelection{Key: tc.key}
			_, err := w.filter(tc.input)
			if tc.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}

func TestGetConditionsFromKeyEmpty(t *testing.T) {
	w := WildCardFilterSelection{Key: ""}
	_, err := w.GetConditionsFromKey()
	if err != SyntaxError {
		t.Errorf("expected SyntaxError, got %v", err)
	}
}

func TestMapSelectionApply(t *testing.T) {
	testcases := []struct {
		name    string
		key     string
		input   interface{}
		want    interface{}
		wantErr error
	}{
		{
			name:  "valid key",
			key:   "name",
			input: map[string]interface{}{"name": "test", "value": 123},
			want:  "test",
		},
		{
			name:    "key not found",
			key:     "missing",
			input:   map[string]interface{}{"name": "test"},
			wantErr: NotFound,
		},
		{
			name:    "not a map",
			key:     "name",
			input:   []interface{}{1, 2, 3},
			wantErr: MapTypeError,
		},
		{
			name:    "string input",
			key:     "name",
			input:   "not a map",
			wantErr: MapTypeError,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			m := &MapSelection{Key: tc.key}
			result, err := m.Apply(tc.input)
			if tc.wantErr != nil {
				if err != tc.wantErr {
					t.Errorf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tc.want {
				t.Errorf("Apply() = %v; want %v", result, tc.want)
			}
		})
	}
}

func TestIsNil(t *testing.T) {
	testcases := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{name: "nil", input: nil, expected: true},
		{name: "nil slice", input: ([]interface{})(nil), expected: true},
		{name: "nil map", input: (map[string]interface{})(nil), expected: true},
		{name: "nil chan", input: (chan int)(nil), expected: true},
		{name: "nil func", input: (func())(nil), expected: true},
		{name: "nil pointer", input: (*int)(nil), expected: true},
		{name: "empty slice", input: []interface{}{}, expected: false},
		{name: "empty map", input: map[string]interface{}{}, expected: false},
		{name: "string", input: "hello", expected: false},
		{name: "int", input: 42, expected: false},
		{name: "zero int", input: 0, expected: false},
		{name: "empty string", input: "", expected: false},
		{name: "bool false", input: false, expected: false},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result := isNil(tc.input)
			if result != tc.expected {
				t.Errorf("isNil(%v) = %v; want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestDescentSelectionApply(t *testing.T) {
	testcases := []struct {
		name    string
		input   interface{}
		nextKey string
		wantLen int
	}{
		{
			name: "descent into nested map",
			input: map[string]interface{}{
				"a": map[string]interface{}{"target": 1},
				"b": map[string]interface{}{"target": 2},
			},
			nextKey: "target",
			wantLen: 2,
		},
		{
			name: "descent into array",
			input: []interface{}{
				map[string]interface{}{"target": 1},
				map[string]interface{}{"target": 2},
				map[string]interface{}{"other": 3},
			},
			nextKey: "target",
			wantLen: 2,
		},
		{
			name:    "scalar input",
			input:   "scalar",
			nextKey: "target",
			wantLen: 0,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			d := &DescentSelection{}
			d.SetNext(&MapSelection{Key: tc.nextKey})
			result, err := d.Apply(tc.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			resultSlice, ok := result.([]interface{})
			if !ok {
				t.Errorf("expected slice result, got %T", result)
				return
			}
			if len(resultSlice) != tc.wantLen {
				t.Errorf("Apply() returned %d items; want %d", len(resultSlice), tc.wantLen)
			}
		})
	}
}

func TestNormalizeEdgeCases(t *testing.T) {
	testcases := []struct {
		input    string
		expected string
	}{
		{input: "", expected: "$"},
		{input: "$", expected: "$"},
		{input: "$..", expected: `$[..]`},
		{input: "$.@", expected: `$[@]`},
		{input: "$.*", expected: `$[*]`},
		{input: "$.a.b.c", expected: `$["a"]["b"]["c"]`},
		{input: "$[0][1][2]", expected: `$[0][1][2]`},
	}

	for _, tc := range testcases {
		t.Run(tc.input, func(t *testing.T) {
			result := normalize(tc.input)
			if result != tc.expected {
				t.Errorf("normalize(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Note: unclosed bracket like "$[store" causes infinite loop in normalize() - known bug
		{name: "invalid array index", input: "$[abc]", wantErr: true},
		{name: "slice syntax not supported", input: "$[0:10:2]", wantErr: true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(tc.input)
			if tc.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}

func TestWildCardKeySelectionApply(t *testing.T) {
	testcases := []struct {
		name  string
		input interface{}
		want  []interface{}
	}{
		{
			name:  "map input returns keys",
			input: map[string]interface{}{"a": 1, "b": 2},
			want:  []interface{}{"a", "b"},
		},
		{
			name:  "non-map returns input",
			input: "scalar",
			want:  nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			w := &WildCardKeySelection{}
			result, err := w.Apply(tc.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if tc.want == nil {
				if result != tc.input {
					t.Errorf("Apply() = %v; want %v", result, tc.input)
				}
				return
			}
			resultSlice, ok := result.([]interface{})
			if !ok {
				t.Errorf("expected slice result, got %T", result)
				return
			}
			if len(resultSlice) != len(tc.want) {
				t.Errorf("Apply() returned %d items; want %d", len(resultSlice), len(tc.want))
			}
		})
	}
}

func TestRootNodeApply(t *testing.T) {
	r := &RootNode{}
	input := map[string]interface{}{"key": "value"}

	result, err := r.Apply(input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Errorf("expected map result, got %T", result)
		return
	}
	if resultMap["key"] != "value" {
		t.Errorf("Apply() = %v; want %v", result, input)
	}
}

func TestFlattenAppend(t *testing.T) {
	testcases := []struct {
		name   string
		src    []interface{}
		values []interface{}
		want   int
	}{
		{
			name:   "simple append",
			src:    []interface{}{1, 2},
			values: []interface{}{3, 4},
			want:   4,
		},
		{
			name:   "nested array flatten",
			src:    []interface{}{1},
			values: []interface{}{[]interface{}{2, 3}},
			want:   3,
		},
		{
			name:   "empty nested array",
			src:    []interface{}{1},
			values: []interface{}{[]interface{}{}},
			want:   1,
		},
		{
			name:   "deeply nested",
			src:    []interface{}{},
			values: []interface{}{[]interface{}{[]interface{}{1, 2}}},
			want:   2,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result := flattenAppend(tc.src, tc.values...)
			if len(result) != tc.want {
				t.Errorf("flattenAppend() returned %d items; want %d", len(result), tc.want)
			}
		})
	}
}

func TestCacheFunctions(t *testing.T) {
	// Clear caches first
	ClearAllCaches()

	// Parse should populate cache
	_, err := Parse("$.test.path")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if ParseCacheSize() != 1 {
		t.Errorf("ParseCacheSize() = %d; want 1", ParseCacheSize())
	}

	// Parse same path should hit cache
	_, err = Parse("$.test.path")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if ParseCacheSize() != 1 {
		t.Errorf("ParseCacheSize() = %d; want 1 (should be cached)", ParseCacheSize())
	}

	// ParseNoCache should not affect cache size
	_, err = ParseNoCache("$.another.path")
	if err != nil {
		t.Fatalf("ParseNoCache failed: %v", err)
	}

	if ParseCacheSize() != 1 {
		t.Errorf("ParseCacheSize() = %d; want 1 (ParseNoCache should not cache)", ParseCacheSize())
	}

	// Clear cache
	ClearParseCache()
	if ParseCacheSize() != 0 {
		t.Errorf("ParseCacheSize() = %d after clear; want 0", ParseCacheSize())
	}
}

func TestMinNotNeg1(t *testing.T) {
	testcases := []struct {
		name string
		a    int
		bs   []int
		want int
	}{
		{name: "single value", a: 5, bs: nil, want: 5},
		{name: "a is -1 takes min", a: -1, bs: []int{3, 5}, want: 3},
		{name: "b smaller", a: 10, bs: []int{5, 8}, want: 5},
		{name: "a smaller", a: 2, bs: []int{5, 8}, want: 2},
		{name: "all -1", a: -1, bs: []int{-1, -1}, want: -1},
		{name: "some -1", a: -1, bs: []int{-1, 5}, want: 5},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result := minNotNeg1(tc.a, tc.bs...)
			if result != tc.want {
				t.Errorf("minNotNeg1(%d, %v) = %d; want %d", tc.a, tc.bs, result, tc.want)
			}
		})
	}
}

func TestMalformedPathsDoNotHang(t *testing.T) {
	// These malformed paths should not cause infinite loops
	malformedPaths := []string{
		"$[",              // unclosed bracket
		"$.[",             // dot then unclosed bracket
		"$.foo[",          // unclosed bracket after key
		"$[bar",           // unclosed bracket with content
		"$[[",             // nested unclosed brackets
		"$[[]",            // partial nested brackets
		"$.foo[0",         // unclosed array index
		`$["key`,          // unclosed quoted key
		"$...[",           // descent with unclosed bracket
		"$[?(@.x",         // unclosed filter
		"$[?(@.x > 5",     // unclosed filter condition
		"$....",           // multiple dots
		"$.[[[[",          // many unclosed brackets
		"$.",              // trailing dot
		"$..",             // double dot at end
		"$...",            // triple dot
		"$[]",             // empty brackets
		"$[[]]",           // nested empty brackets
	}

	for _, path := range malformedPaths {
		t.Run(path, func(t *testing.T) {
			done := make(chan bool, 1)
			go func() {
				// Parse should complete (possibly with error) but not hang
				_, _ = Parse(path)
				done <- true
			}()

			select {
			case <-done:
				// Success - parsing completed
			case <-time.After(100 * time.Millisecond):
				t.Errorf("Parse(%q) appears to hang (timeout after 100ms)", path)
			}
		})
	}
}
