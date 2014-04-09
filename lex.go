package jsonpath

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const eof = -1

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
