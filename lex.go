package jsonpath

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const eof = -1

type tokeType int8

const (
	LexError tokeType = iota
	Start
	HashKey
	ArrayKey
	Splat
	Recurse
	Filter
	Map
)

type lex struct {
	path  string
	start int // the start position in the sring
	pos   int // the current position in the string
	width int // the width of the string
}

type Token interface {
	Type() tokeType
	Tok() string
}
type StringToken struct {
	token string
}
type HashKeyToken struct {
	StringToken
}
type ArrayKeyToken struct {
	StringToken
}
type FilterToken struct {
	StringToken
}
type MapToken struct {
	StringToken
}
type SplatToken struct{}
type StartToken struct{}
type RecurseToken struct{}
type LexErrorToken struct {
	StringToken
	Error string
}

func (s *StringToken) Tok() string {
	return s.token
}
func (_ *ArrayKeyToken) Type() tokeType {
	return ArrayKey
}
func (_ *HashKeyToken) Type() tokeType {
	return HashKey
}
func (_ *MapToken) Type() tokeType {
	return Map
}
func (_ *FilterToken) Type() tokeType {
	return Filter
}
func (_ *SplatToken) Tok() string {
	return "*"
}
func (_ *SplatToken) Type() tokeType {
	return Splat
}
func (_ *RecurseToken) Tok() string {
	return ".."
}
func (_ *RecurseToken) Type() tokeType {
	return Recurse
}
func (_ *LexErrorToken) Type() tokeType {
	return LexError
}
func (_ *StartToken) Tok() string {
	return "$"
}
func (_ *StartToken) Type() tokeType {
	return Start
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

func (l *lex) peek() (r rune) {
	r = l.next()
	l.back()
	return
}
func (l *lex) skipTo(s rune) (b bool) {
	r := l.next()
	for ; r != s && r != eof; r = l.next() {
		b = true
	}
	l.back()
	return
}

func (l *lex) skipToRun(s string) (b bool) {
	r := l.next()
	for ; !strings.ContainsRune(s, r) && r != eof; r = l.next() {
		b = true
	}
	if r != eof {
		l.back()
	}
	return
}
func (l *lex) accept(s rune) bool {
	if r := l.next(); r == s {
		return true
	}
	l.back()
	return false
}
func (l *lex) ignore(s rune) bool {
	if l.accept(s) {
		l.collopse()
		return true
	}
	return false
}
func (l *lex) isAtEnd() bool {
	return l.pos >= len(l.path)
}
func (l *lex) acceptRun(s string) bool {
	pos := l.pos
	for _, r := range s {
		if !l.accept(r) {
			l.pos = pos // reset back
			return false
		}
	}
	return true
}
func (l *lex) debug() string {
	return fmt.Sprintf("path:%q (%q) start: %d pos: %d", l.path[l.start:], l.path[l.start:l.pos], l.start, l.pos)
}
func (l *lex) acceptAny(s string) bool {
	for _, r := range s {
		if l.accept(r) {
			return true
		}
	}
	return false
}
func (l *lex) acceptAnyRun(s string) (b bool) {
	for r := l.next(); strings.ContainsRune(s, r); r = l.next() {
		b = true
	}
	l.back()
	return
}
func (l *lex) collopse() string {
	s, e := l.start, l.pos
	l.start = l.pos
	return l.path[s:e]
}
func (l *lex) tokeLen() int {
	return l.pos - l.start
}
