package jsonpath

import (
	"errors"
	"io"
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
)

func applyNext(nn node, v interface{}) (interface{}, error) {
	if nn == nil {
		return v, nil
	}
	return nn.Apply(v)
}

// A root type just has the NextNode which is it's self. If node is nil, we
// Just return the interface, there is no need to filter.
type RootNode struct {
	NextNode node
}

func (r *RootNode) SetNext(n node) {
	if r.NextNode == nil {
		r.NextNode = n
	}
	r.NextNode.SetNext(n)
}

func (r *RootNode) Apply(v interface{}) (interface{}, error) {
	return applyNext(r.NextNode, v)
}

type MapSelection struct {
	Key string
	RootNode
}

func (m *MapSelection) Apply(v interface{}) (interface{}, error) {
	mv, ok := v.(map[string]interface{})
	if !ok {
		return v, MapTypeError
	}
	return applyNext(m.NextNode, mv[m.Key])
}

type ArraySelection struct {
	Key int
	RootNode
}

func (a *ArraySelection) Apply(v interface{}) (interface{}, error) {
	arv, ok := v.([]interface{})
	if !ok {
		return v, ArrayTypeError
	}
	return applyNext(a.NextNode, arv[a.Key])
}

type WildCardSelection struct {
	RootNode
}

func (w *WildCardSelection) Apply(v interface{}) (interface{}, error) {
	switch tv := v.(type) {
	case map[string]interface{}:
		var ret []interface{}
		for _, val := range tv {
			rval, err := applyNext(w.NextNode, val)
			// Don't add anything that causes an error or returns nil.
			if err == nil || rval != nil {
				ret = append(ret, rval)
			}
		}
		return ret, nil
	case []interface{}:
		var ret []interface{}
		for _, val := range tv {
			rval, err := applyNext(w.NextNode, val)
			// Don't add anything that causes an error or returns nil.
			if err == nil || rval != nil {
				ret = append(ret, rval)
			}
		}
		return ret, nil

	default:
		return applyNext(w.NextNode, v)
	}
}

type DescentSelection struct {
	RootNode
}

func (d *DescentSelection) Apply(v interface{}) (interface{}, error) {
	var ret []interface{}
	rval, err := applyNext(d.NextNode, v)
	if err != nil {
		return nil, err
	}
	if rval != nil {
		ret = append(ret, rval)
	}
	switch tv := v.(type) {
	default:
		return ret, nil
	case map[string]interface{}:
		for _, val := range tv {
			rval, err := d.Apply(val)
			// Don't add anything that causes an error or returns nil.
			if err == nil || rval != nil {
				ret = append(ret, rval)
			}
		}
		return ret, nil
	case []interface{}:
		for _, val := range tv {
			rval, err := d.Apply(val)
			// Don't add anything that causes an error or returns nil.
			if err == nil || rval != nil {
				ret = append(ret, rval)
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
		if s[0] == '*' {
			r += "[*]"
			if len(s) == 1 {
				break
			}
			s = s[1:]
		}

		n := minNotNeg1(strings.Index(s, "["), strings.Index(s, "."))
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
	case "[.":
		return &DescentSelection{}, rs, nil
	case "[?", "[(":
		return nil, rs, NotSupportedError
	default: // Assume it's a array index otherwise.
		i, err := strconv.Atoi(s[1:n])
		if err != nil {
			return nil, rs, SyntaxError
		}
		return &ArraySelection{Key: i}, rs, nil
	}
}

func Parse(s string) (Applicator, error) {
	var nn node
	var err error
	s = normalize(s)
	rt := RootNode{}
	// Remove the starting '$'
	s = s[1:]
	for len(s) > 0 {
		nn, s, err = getNode(s)
		if err != nil {
			return nil, err
		}
		rt.SetNext(nn)
	}
	return &rt, nil
}
