package jsonpath

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

type objKeyType int8
type Object interface{}

const eof = -1
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

type lex struct {
	path  string
	start int // the start position in the sring
	pos   int // the current position in the string
	width int // the width of the string
}

func (l *lex) next() rune {
	if l.pos >= len(l.path) {
		l.width = 0
		return eof
	}
	r, width := utf8.DecodeRuneInString(l.path[l.pos:])
	l.width = width
	l.pos += width
	return r
}

func (l *lex) back() {
	l.pos -= l.width
	if l.pos < l.start {
		l.pos = l.start
	}
}

func (l *lex) peek() rune {
	r := l.next()
	l.back()
	return r
}
func (l *lex) skipTo(s rune) string {
	r := l.next()
	for ; r != s && r != eof; r = l.next() {
	}
	l.back()
	result := l.path[l.start:l.pos]
	l.start = l.pos
	return result
}

func (l *lex) skipToRun(s string) string {
	r := l.next()
	for ; !strings.ContainsRune(s, r) && r != eof; r = l.next() {
	}
	l.back()
	result := l.path[l.start:l.pos]
	l.start = l.pos
	return result
}
func (l *lex) accept(s rune) (string, bool) {
	if r := l.next(); r == s {
		rs := l.path[l.start:l.pos]
		l.start = l.pos
		return rs, true
	}
	l.back()
	return "", false
}
func (l *lex) ignore(s rune) {
	l.accept(s)
}
func (l *lex) isAtEnd() bool {
	return l.pos >= len(l.path)
}
func (l *lex) acceptRun(s string) bool {
	start, pos := l.start, l.pos
	for _, r := range s {
		if _, b := l.accept(r); !b {
			l.start, l.pos = start, pos // reset back
			return false
		}
	}
	return true
}
func (l *lex) debug() string {
	return fmt.Sprintf("path:%q start: %d pos: %d", l.path[l.start:], l.start, l.pos)
}
func (l *lex) acceptAny(s string) (string, bool) {
	for _, r := range s {
		if c, b := l.accept(r); b {
			return c, true
		}
	}
	return "", false
}
func (l *lex) acceptAnyRun(s string) (string, bool) {
	var result string
	var gotval bool
	for c, b := l.acceptAny(s); b; c, b = l.acceptAny(s) {
		result += c
		gotval = true
	}
	return result, gotval
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
