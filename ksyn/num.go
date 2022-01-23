package ksyn

import (
	"fmt"
)

type Num struct {
	Num  string
	Unit string

	start, end Pos
}

func (n Num) Start() Pos {
	return n.start
}

func (n Num) End() Pos {
	return n.end
}

func ch_is_num(ch byte) bool {
	if ch >= '0' && ch <= '9' {
		return true
	}
	return false
}

func (p *parser) num() (Expr, error) {
	start := p.pos

	str := p.str()

	ch := p.ch()
	if !ch_is_num(ch) {
		return nil, p.unexpectedRune("digit")
	}

	v := 0

	for ch_is_num(ch) {
		v++
		p.next()
		ch = p.ch()
	}

	if ch == '.' {
		v++
		p.next()
		ch = p.ch()

		for ch_is_num(ch) {
			v++
			p.next()
			ch = p.ch()
		}
	}

	// with a unit?
	uvstr := p.str()
	uv := 0
	if ch_is_ident(ch) {
		uv++
		p.next()
		ch = p.ch()
		for ch_is_ident_rest(ch) {
			uv++
			p.next()
			ch = p.ch()
		}
	}

	end := p.pos

	return Num{
		Num:  str[:v],
		Unit: uvstr[:uv],

		start: start,
		end:   end,
	}, nil
}

func (n Num) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	if n.Unit == "" {
		r += fmt.Sprintf("%T %#v\n", n, n.Num)
	} else {
		r += fmt.Sprintf("%T %#v %#v\n", n, n.Num, n.Unit)
	}
	return r
}
