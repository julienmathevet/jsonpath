package jsonpath

import "strings"

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

func normalize(s string) (string, error) {
	if s == "" || s == "$." {
		return "$", nil
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
	return r, nil
}
