package jsonpath

import (
	"errors"
	"fmt"
	"go/token"
	"go/types"
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
	re1 := regexp.MustCompile(`(\s+\|\|\s+)`)
	// split by || condition
	orConditions := re1.Split(w.Key, -1)
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
		//re, err := regexp.Compile(`[\S]+`)
		re, err := regexp.Compile(`([@$.\w]+) ([<=>!~]{1,2}) ([']?[\w\W\d\s]+[']?)`)
		if err != nil {
			return val, err
		}
		simpleRe, err := regexp.Compile(`([@$.\w]+)$`)
		if err != nil {
			return val, err
		}

		//ops := re.FindAllString(w.Key, -1)
		match := re.FindAllStringSubmatch(condition, -1)

		if match == nil || len(match) == 0 {
			match = simpleRe.FindAllStringSubmatch(condition, -1)
			if match == nil || len(match) == 0 {
				return val, SyntaxError
			}
		}

		wa, _ := Parse(strings.Replace(match[0][1], "@", "$", 1))
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
	r := "$"
	// first thing to do is read passed the $
	if s[0] == '$' {
		s = s[1:]
	}
	for len(s) > 0 {

		// Grab all the bracketed entries
		for len(s) > 0 && s[0] == '[' {
			n := strings.Index(s, "]")
			r += s[0 : n+1]
			s = s[n+1:]
		}
		if len(s) <= 0 || s == "." {
			break
		}

		if s[0] == '.' {
			if s[1] == '.' {
				r += "[..]"
				s = s[2:]

			} else {
				s = s[1:]
			}
		}
		if len(s) == 0 {
			break
		}
		if s[0] == '*' {
			r += "[*]"
			if len(s) == 1 {
				break
			}
			s = s[1:]
		}
		if s[0] == '@' {
			r += "[@]"
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
			r += `["` + s[:n] + `"]`
			s = s[n:]
		} else {
			r += `["` + s + `"]`
			s = ""
		}
	}
	return r
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
	switch obj1.(type) {
	case string:
		sobj1 = strings.ReplaceAll(obj1.(string), "'", "")
		sobj1 = fmt.Sprintf("%v", sobj1)
	default:
		sobj1 = fmt.Sprintf("%v", obj1)
	}
	var sobj2 string
	switch obj1.(type) {
	case string:
		sobj2 = strings.ReplaceAll(obj2.(string), "'", "")
		sobj2 = fmt.Sprintf("^%v$", sobj2)
	default:
		sobj2 = fmt.Sprintf("^%v$", obj2)
	}
	re, err := regexp.Compile(sobj2)
	if err != nil {
		return false, err
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
	var sobj1 string
	switch obj1.(type) {
	case string:
		sobj1 = strings.ReplaceAll(obj1.(string), "'", "")
		sobj1 = strings.ReplaceAll(sobj1, "\"", "")
		sobj1 = strings.ReplaceAll(sobj1, "\\", "")
		sobj1 = strings.ReplaceAll(sobj1, ".", "")
		sobj1 = fmt.Sprintf("\"%v\"", sobj1)
	default:
		sobj1 = fmt.Sprintf("%v", obj1)
	}
	var sobj2 string
	switch obj1.(type) {
	case string:
		sobj2 = strings.ReplaceAll(obj2.(string), "'", "")
		sobj2 = fmt.Sprintf("\"%v\"", sobj2)
	default:
		sobj2 = fmt.Sprintf("%v", obj2)
	}

	exp := fmt.Sprintf("%v %s %v", sobj1, op, sobj2)
	fset := token.NewFileSet()
	res, err := types.Eval(fset, nil, 0, exp)
	if err != nil {
		return false, err
	}
	if res.IsValue() == false || (res.Value.String() != "false" && res.Value.String() != "true") {
		return false, fmt.Errorf("result should only be true or false")
	}
	if res.Value.String() == "true" {
		return true, nil
	}
	return false, nil
}
