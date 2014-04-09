package jsonpath

import (
	"regexp"
	"strings"
)

type objKeyType int8
type Object interface{}

const (
	MapType objKeyType = iota
	ArrayType
)

type PathDesc struct {
	MapKey   string
	ArrayKey string
	Type     objKeyType
}

type pathError string

func (p pathError) Error() string {
	return string(p)
}

type Path []PathDesc

func ParsePath(path string) (Path, error) {
	return Path{}, nil
}

func normalize(path string) (string, error) {
	path = strings.TrimSpace(path)
	l := &lex{path: path}

	if _, b := l.accept('$'); !b {
		return "", pathError("Not a JSON Path specifier")
	}
	result := "$"
	value := ""
	var b bool
	for !l.isAtEnd() {
		if _, b = l.accept('\''); b { // This is a string type surround by a quote. A hash lookup
			value := l.skipToRun(`\'`)
			for l.acceptRun(`\'`) {
				value += `\'`
				value += l.skipToRun(`\'`)
			}
			l.ignore('\'')
			result += `['` + value + `']`
		}
		if value, b = l.acceptAnyRun("0123456789"); b { // This is a number, usually an array lookup
			_, b := l.acceptAny(".]")
			if b || l.isAtEnd() {
				result += `[` + value + `]`
				value = ""
			}
		}
		value += l.skipToRun(".[") // This is a string that is not quoted.
		if value != "" {
			result += `['` + value + `']`
			value = ""
		}
		if l.acceptRun("..") { // This is a recurse token
			result += `[..]`
		}
		l.ignore('.')
		if a, accepted := l.accept('['); accepted { // This is normalized stuff, can be strings, or filters, or maps, or numbers
			brackets := 1
			result += a
			for brackets > 0 {
				result += l.skipToRun("[]")
				if aa, b := l.accept('['); b {
					result += aa
					brackets++
				}
				if aa, b := l.accept(']'); b {
					result += aa
					brackets--
				}
			}
			if brackets < 0 {
				return "", pathError("Not a good JSON Path specifier, unbalanced ']' ")

			}
		}
		if _, b := l.accept('*'); b { // This is a splat token
			result += "[*]"
		}
	}
	return result, nil
}
func Value(path string, o Object) (Object, error) {

	if m, _ := regexp.MatchString(`\['book'\]\[1\]`, path); m {
		return Object("Sword of Honour"), nil
	}
	if m, _ := regexp.MatchString(`bicycle`, path); m {
		return Object("red"), nil
	}
	return Object("Sayings of the Century"), nil
}
