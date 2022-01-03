package ksyn

import (
	"fmt"
	"strings"
)

var reserved = func() map[string]bool {
	list := `
break        default      func         interface    select
case         defer        go           map          struct
chan         else         goto         package      switch
const        fallthrough  if           range        type
continue     for          import       return       var

grep
`
	r := map[string]bool{}
	for _, v := range strings.Fields(list) {
		r[v] = true
	}
	return r
}()

type Ident struct {
	Ident string

	start, end Pos
}

func (i Ident) Start() Pos {
	return i.start
}
func (i Ident) End() Pos {
	return i.end
}

func ch_is_ident(ch byte) bool {
	if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' {
		return true
	}
	return false
}
func ch_is_ident_rest(ch byte) bool {
	if ch_is_ident(ch) || ch_is_num(ch) {
		return true
	}
	return false
}

func (p *parser) peekIdent() bool {
	s := p.str()
	return peekIdentStr(s)
}
func peekIdentStr(s string) bool {
	if len(s) == 0 {
		return false
	}
	l := 0
	if !ch_is_ident(s[l]) {
		return false
	}
	l++
	for l < len(s) {
		if !ch_is_ident_rest(s[l]) {
			break
		}
		l++
	}
	id := s[:l]
	if reserved[id] {
		return false
	}
	return true
}

func (p *parser) ident() (Ident, error) {
	s := p.str()
	l := 0
	start := p.pos

	ch := p.ch()
	if !ch_is_ident(ch) {
		return Ident{}, p.unexpected("identifier")
	}
	p.next()
	l++

	for ch_is_ident_rest(p.ch()) {
		p.next()
		l++
	}

	end := p.pos

	id := s[:l]

	if reserved[id] {
		return Ident{}, Error{
			Pos: start,
			Err: fmt.Errorf("Invalid identifier: %#v is a reserved word", id),
		}
	}

	return Ident{
		Ident: id,

		start: start,
		end:   end,
	}, nil
}

func (i Ident) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T %#v\n", i, i.Ident)
	return r
}
