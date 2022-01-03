package ksyn

import (
	"fmt"
)

type Int struct {
	Int int

	start, end Pos
}

func (i Int) Start() Pos {
	return i.start
}

func (i Int) End() Pos {
	return i.end
}

func ch_is_num(ch byte) bool {
	if ch >= '0' && ch <= '9' {
		return true
	}
	return false
}

func (p *parser) num() (Expr, error) {
	start := p.pos

	ch := p.ch()
	if !ch_is_num(ch) {
		return nil, p.unexpectedRune("digit")
	}

	v := 0

	for ch_is_num(ch) {
		v *= 10
		v += int(ch - '0')
		p.next()
		ch = p.ch()
	}

	end := p.pos

	return Int{
		Int: v,

		start: start,
		end:   end,
	}, nil
}

func (i Int) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T %d\n", i, i.Int)
	return r
}
