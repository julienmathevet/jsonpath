package jsonpath

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Applicator interface {
	Apply(v interface{}) (interface{}, error)
}

type node interface {
	Applicator
	SetNext(v node)
}

var (
	MapTypeError      = errors.New("Expected Type to be a Map.")
	ArrayTypeError    = errors.New("Expected Type to be an Array.")
	SyntaxError       = errors.New("Bad Syntax.")
	NotSupportedError = errors.New("Not Supported")
	NotFound          = errors.New("Not Found")
	IndexOutOfBounds  = errors.New("Out of Bounds")
)

// Pre-compiled regexes for filter operations
var (
	filterConditionRe = regexp.MustCompile(`([@$.\w]+) ([<=>!~]{1,2}) ([']?[\w\W\d\s]+[']?)`)
	simpleConditionRe = regexp.MustCompile(`([@$.\w]+)$`)
	orConditionRe     = regexp.MustCompile(`(\s+\|\|\s+)`)
)

// Cache for compiled wildcard patterns
var wildcardCache = make(map[string]*regexp.Regexp)

func applyNext(nn node, v interface{}) (interface{}, error) {
	if nn == nil {
		return v, nil
	}
	return nn.Apply(v)
}

// RootNode is always the top node. It does not really do anything other then
// delegate to the next node, and acts as a starting point.
// Every other node type embeds this node To get the NextNode functions.
type RootNode struct {
	// The next Node in the sequence.
	NextNode node
}

// Set the next node.
func (r *RootNode) SetNext(n node) {
	r.NextNode = n
}

// Apply is the main workhourse, each node type will apply it's filtering rules
// to the provided value, returning the filtered result.
// It is expected that the node will call's it's Next Nodes Apply method as
// need by the rules of the Node.
func (r *RootNode) Apply(v interface{}) (interface{}, error) {
	return applyNext(r.NextNode, v)
}

// MapSelection is a the basic filter for a Map type key. It will look at the
// in coming v and try to turn it into a map[string]interface{} value. If it
// successeeds it will then apply the NextNode to that interface value, or it
// will return the value if it has no NextNode.
type MapSelection struct {
	Key string
	RootNode
}

func (m *MapSelection) Apply(v interface{}) (interface{}, error) {
	mv, ok := v.(map[string]interface{})
	if !ok {
		return v, MapTypeError
	}
	nv, ok := mv[m.Key]
	if !ok {
		return nil, NotFound
	}
	return applyNext(m.NextNode, nv)
}

// ArrySelection is a the basic filter for a Array type key. It is like MapSelection but for Arrays.
type ArraySelection struct {
	Key int
	RootNode
}

func (a *ArraySelection) Apply(v interface{}) (interface{}, error) {
	arv, ok := v.([]interface{})
	if !ok {
		return v, ArrayTypeError
	}
	// Check to see if the value is in bounds for the array.
	if a.Key < 0 || a.Key >= len(arv) {
		return nil, IndexOutOfBounds

	}
	return applyNext(a.NextNode, arv[a.Key])
}

// WildCardSelection is a filter that grabs all the values and returns an Array of them
// It applies it's NextNode on each value.
type WildCardSelection struct {
	RootNode
}

func (w *WildCardSelection) Apply(v interface{}) (interface{}, error) {
	switch tv := v.(type) {
	case map[string]interface{}:
		var ret []interface{}
		var keys []string
		for key := range tv {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			rval, err := applyNext(w.NextNode, tv[key])
			// Don't add anything that causes an error or returns nil.
			if err == nil || rval != nil {
				ret = flattenAppend(ret, rval)
			}
		}
		return ret, nil
	case []interface{}:
		var ret []interface{}
		for _, val := range tv {
			rval, err := applyNext(w.NextNode, val)
			// Don't add anything that causes an error or returns nil.
			if err == nil || rval != nil {
				ret = flattenAppend(ret, rval)
			}
		}
		return ret, nil

	default:
		return applyNext(w.NextNode, v)
	}
}

type WildCardKeySelection struct {
	RootNode
}

func (w *WildCardKeySelection) Apply(v interface{}) (interface{}, error) {
	switch tv := v.(type) {
	case map[string]interface{}:
		var ret []interface{}
		var keys []string
		for key, _ := range tv {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			rval, err := applyNext(w.NextNode, key)
			// Don't add anything that causes an error or returns nil.
			if err == nil || rval != nil {
				ret = flattenAppend(ret, rval)
			}
		}
		return ret, nil

	default:
		return applyNext(w.NextNode, v)
	}
}

type WildCardFilterSelection struct {
	RootNode
	Key string
	// Cache for parsed sub-paths to avoid re-parsing on each filter call
	pathCache map[string]Applicator
}

func (w *WildCardFilterSelection) Apply(v interface{}) (interface{}, error) {
	var ret []interface{}

	switch arv := v.(type) {
	case map[string]interface{}:
		rval, err := w.filter(arv)
		if err == nil && rval != nil {
			ret = append(ret, rval)
		}
	case []interface{}:
		for _, val := range arv {
			rval, err := w.filter(val)
			// Don't add anything that causes an error or returns nil.
			if err == nil && rval != nil {
				ret = append(ret, rval)
			}
		}

	default:

		return v, ArrayTypeError
	}
	return ret, nil
}

func (w *WildCardFilterSelection) GetConditionsFromKey() ([]string, error) {
	if w.Key == "" {
		return nil, SyntaxError
	}
	conditions := []string{}
	// split by || condition using pre-compiled regex
	orConditions := orConditionRe.Split(w.Key, -1)
	for _, orCondition := range orConditions {
		// if the orCondition contains an && condition and the terms of the end condition are between parentheses
		// append to conditions
		conditions = append(conditions, orCondition)
	}
	return conditions, nil
}

func (w *WildCardFilterSelection) filter(val interface{}) (interface{}, error) {
	_, ok := val.(map[string]interface{})
	if !ok {
		return val, MapTypeError
	}

	conditions, err := w.GetConditionsFromKey()
	if err != nil {
		return nil, err
	}
	shouldKeep := false
	for _, condition := range conditions {
		// Use pre-compiled regexes
		match := filterConditionRe.FindAllStringSubmatch(condition, -1)

		if match == nil || len(match) == 0 {
			match = simpleConditionRe.FindAllStringSubmatch(condition, -1)
			if match == nil || len(match) == 0 {
				return val, SyntaxError
			}
		}

		// Use cached parsed path or parse and cache it
		pathExpr := match[0][1]
		wa := w.getCachedPath(pathExpr)
		subv, _ := wa.Apply(val)
		if subv == nil {
			continue
		}
		if len(match[0]) == 2 {
			shouldKeep = true
		} else if len(match[0]) == 4 {
			op := match[0][2]
			if op == "=~" || op == "!~" {
				isOk, _ := cmp_wildcard(subv, match[0][3], op)
				if !isOk {
					continue
				} else {
					shouldKeep = true
				}
			}
			isOk, _ := cmp_any(subv, match[0][3], op)
			if !isOk {
				continue
			} else {
				shouldKeep = true
			}
		}
	}
	if !shouldKeep {
		return nil, nil
	}
	rval, err := applyNext(w.NextNode, val)
	return rval, err
}

// getCachedPath returns a cached parsed path or parses and caches it
func (w *WildCardFilterSelection) getCachedPath(pathExpr string) Applicator {
	if w.pathCache == nil {
		w.pathCache = make(map[string]Applicator)
	}
	if cached, ok := w.pathCache[pathExpr]; ok {
		return cached
	}
	parsed, _ := Parse(strings.Replace(pathExpr, "@", "$", 1))
	w.pathCache[pathExpr] = parsed
	return parsed
}

// DescentSelection is a filter that recursively descends applying it's NextNode and
// corrlating the results.
type DescentSelection struct {
	RootNode
}

func isNil(i interface{}) bool {
	vi := reflect.ValueOf(i)
	if !vi.IsValid() {
		return true
	}
	switch vi.Kind() {
	case reflect.Chan, reflect.Interface, reflect.Func, reflect.Slice, reflect.Map, reflect.Ptr:
		return vi.IsNil()
	default:
		return false
	}
}
func flattenAppend(src []interface{}, values ...interface{}) []interface{} {
	for _, value := range values {
		av, ok := value.([]interface{})
		if ok {
			if len(av) > 0 {
				src = flattenAppend(src, av...)
			}
		} else {
			src = append(src, value)
		}
	}
	return src
}

func (d *DescentSelection) Apply(v interface{}) (interface{}, error) {
	var ret []interface{}
	rval, err := applyNext(d.NextNode, v)

	// Ignore errors here.
	if err == nil && !isNil(rval) {
		ret = flattenAppend(ret, rval)
	}
	switch tv := v.(type) {
	default:
		return ret, nil
	case map[string]interface{}:
		var keys []string
		for key, _ := range tv {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			rval, err := d.Apply(tv[key])
			// Don't add anything that causes an error or returns nil.
			if err == nil && !isNil(rval) {
				ret = flattenAppend(ret, rval)
			}
		}
		return ret, nil
	case []interface{}:
		for _, val := range tv {
			rval, err := d.Apply(val)
			// Don't add anything that causes an error or returns nil.
			if err == nil && !isNil(rval) {
				ret = flattenAppend(ret, rval)
			}
		}
		return ret, nil
	}
}

// The first thing we need to do is transform a dot–notation to a bracket–notation.

func minNotNeg1(a int, bs ...int) int {
	m := a
	for _, b := range bs {
		if a == -1 || (b != -1 && b < m) {
			m = b
		}
	}
	return m
}

func normalize(s string) string {
	if s == "" || s == "$." {
		return "$"
	}

	// Pre-allocate builder with estimated capacity
	var b strings.Builder
	b.Grow(len(s) * 2)
	b.WriteByte('$')

	// first thing to do is read passed the $
	if s[0] == '$' {
		s = s[1:]
	}
	for len(s) > 0 {

		// Grab all the bracketed entries
		for len(s) > 0 && s[0] == '[' {
			n := strings.Index(s, "]")
			b.WriteString(s[0 : n+1])
			s = s[n+1:]
		}
		if len(s) <= 0 || s == "." {
			break
		}

		if s[0] == '.' {
			if s[1] == '.' {
				b.WriteString("[..]")
				s = s[2:]

			} else {
				s = s[1:]
			}
		}
		if len(s) == 0 {
			break
		}
		if s[0] == '*' {
			b.WriteString("[*]")
			if len(s) == 1 {
				break
			}
			s = s[1:]
		}
		if s[0] == '@' {
			b.WriteString("[@]")
			if len(s) == 1 {
				break
			}
			s = s[1:]
		}

		n := minNotNeg1(strings.Index(s, "["), strings.Index(s, "."))
		if n == 0 {
			continue
		}
		if n != -1 {
			b.WriteString(`["`)
			b.WriteString(s[:n])
			b.WriteString(`"]`)
			s = s[n:]
		} else {
			b.WriteString(`["`)
			b.WriteString(s)
			b.WriteString(`"]`)
			s = ""
		}
	}
	return b.String()
}

func getNode(s string) (node, string, error) {
	var rs string
	if len(s) == 0 {
		return nil, s, io.EOF
	}
	n := strings.Index(s, "]")
	if n == -1 {
		return nil, s, SyntaxError
	}
	if len(s) > n {
		rs = s[n+1:]
	}
	switch s[:2] {
	case "[\"":
		return &MapSelection{Key: s[2 : n-1]}, rs, nil
	case "[*":
		return &WildCardSelection{}, rs, nil
	case "[@":
		return &WildCardKeySelection{}, rs, nil
	case "[.":
		return &DescentSelection{}, rs, nil
	case "[?", "[(":
		return &WildCardFilterSelection{Key: s[3 : n-1]}, rs, nil
	default: // Assume it's a array index otherwise.
		i, err := strconv.Atoi(s[1:n])
		if err != nil {
			return nil, rs, SyntaxError
		}
		return &ArraySelection{Key: i}, rs, nil
	}
}

// Parse parses the JSONPath and returns a object that can be applied to
// a structure to filter it down.
func Parse(s string) (Applicator, error) {
	var nn node
	var err error
	s = normalize(s)
	rt := RootNode{}
	// Remove the starting '$'
	s = s[1:]
	var c node
	c = &rt
	for len(s) > 0 {
		nn, s, err = getNode(s)
		if err != nil {
			return nil, err
		}
		c.SetNext(nn)
		c = nn
	}
	return &rt, nil
}

func cmp_wildcard(obj1, obj2 interface{}, op string) (bool, error) {
	var sobj1 string
	switch v := obj1.(type) {
	case string:
		sobj1 = strings.ReplaceAll(v, "'", "")
	default:
		sobj1 = fmt.Sprintf("%v", obj1)
	}

	// Build pattern string
	var pattern string
	switch obj1.(type) {
	case string:
		pattern = "^" + strings.ReplaceAll(obj2.(string), "'", "") + "$"
	default:
		pattern = fmt.Sprintf("^%v$", obj2)
	}

	// Use cached regex or compile and cache
	re, ok := wildcardCache[pattern]
	if !ok {
		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			return false, err
		}
		wildcardCache[pattern] = re
	}
	switch op {
	case "=~":
		return re.MatchString(sobj1), nil
	case "!~":
		return !re.MatchString(sobj1), nil
	}
	return false, SyntaxError
}

func cmp_any(obj1, obj2 interface{}, op string) (bool, error) {
	switch op {
	case "<", "<=", "==", ">=", ">", "!=":
	default:
		return false, fmt.Errorf("op should only be <, <=, ==, !=, >= and >")
	}

	// Convert obj2 (from JSON path expression) to comparable value
	obj2Str := strings.ReplaceAll(fmt.Sprintf("%v", obj2), "'", "")

	// Try to compare based on obj1's type
	switch v1 := obj1.(type) {
	case float64:
		v2, err := strconv.ParseFloat(obj2Str, 64)
		if err != nil {
			return false, err
		}
		return compareFloat64(v1, v2, op), nil

	case int:
		v2, err := strconv.ParseInt(obj2Str, 10, 64)
		if err != nil {
			// Try as float
			v2f, err := strconv.ParseFloat(obj2Str, 64)
			if err != nil {
				return false, err
			}
			return compareFloat64(float64(v1), v2f, op), nil
		}
		return compareInt64(int64(v1), v2, op), nil

	case string:
		v2 := strings.ReplaceAll(obj2Str, "'", "")
		return compareString(v1, v2, op), nil

	case bool:
		v2, err := strconv.ParseBool(obj2Str)
		if err != nil {
			return false, err
		}
		return compareBool(v1, v2, op), nil

	default:
		// Fallback: compare string representations
		v1Str := fmt.Sprintf("%v", obj1)
		return compareString(v1Str, obj2Str, op), nil
	}
}

func compareFloat64(a, b float64, op string) bool {
	switch op {
	case "<":
		return a < b
	case "<=":
		return a <= b
	case "==":
		return a == b
	case ">=":
		return a >= b
	case ">":
		return a > b
	case "!=":
		return a != b
	}
	return false
}

func compareInt64(a, b int64, op string) bool {
	switch op {
	case "<":
		return a < b
	case "<=":
		return a <= b
	case "==":
		return a == b
	case ">=":
		return a >= b
	case ">":
		return a > b
	case "!=":
		return a != b
	}
	return false
}

func compareString(a, b string, op string) bool {
	switch op {
	case "<":
		return a < b
	case "<=":
		return a <= b
	case "==":
		return a == b
	case ">=":
		return a >= b
	case ">":
		return a > b
	case "!=":
		return a != b
	}
	return false
}

func compareBool(a, b bool, op string) bool {
	switch op {
	case "==":
		return a == b
	case "!=":
		return a != b
	}
	return false
}
