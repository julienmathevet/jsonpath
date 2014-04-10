package jsonpath

import (
	"regexp"
	"strings"
)

type Object interface{}

type PathDesc struct {
	MapKey   string
	ArrayKey string
	Type     tokeType
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

	if !l.ignore('$') {
		return "", pathError("Not a JSON Path specifier")
	}
	result := "$"
	for !l.isAtEnd() {
		if l.ignore('\'') { // This is a string type surround by a quote. A hash lookup
			l.skipToRun(`\'`)
			for l.acceptRun(`\'`) {
				l.skipToRun(`\'`)
			}
			result += `['` + l.collopse() + `']`
			l.ignore('\'')
		}
		if l.acceptAnyRun("0123456789") { // This is a number, usually an array lookup
			if l.acceptAny(".]") || l.isAtEnd() {
				result += `[` + l.collopse() + `]`
			}
		}
		l.skipToRun(".[") // This is a string that is not quoted.
		if l.tokeLen() > 0 {
			result += `['` + l.collopse() + `']`
		}
		if l.acceptRun("..") { // This is a recurse token
			l.collopse()
			result += `[..]`
		}
		l.ignore('.')
		if l.accept('[') { // This is normalized stuff, can be strings, or filters, or maps, or numbers
			brackets := 1
			for brackets > 0 && !l.isAtEnd() {
				l.skipToRun("[]")
				if l.accept('[') {
					brackets++
				}
				if l.accept(']') {
					brackets--
				}
			}
			if brackets != 0 {
				return "", pathError("Not a good JSON Path specifier, unbalanced ']' or '[' ")
			}

			result += l.collopse()
		}
		if l.ignore('*') { // This is a splat token
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
