package ksyn

import (
	"fmt"
)

type List struct {
	Entries    []Expr
	start, end Pos
}

func (p *parser) exprList() (List, error) {
	start := p.pos

	if err := p.consumeRune('['); err != nil {
		return List{}, err
	}

	r := List{
		start: start,
	}

	for {
		p.chomp()
		if p.ch() == ']' {
			p.next()
			break
		}

		v, err := p.expr()
		if err != nil {
			return List{}, err
		}

		r.Entries = append(r.Entries, v)
	}

	return r, nil
}

func (l List) Start() Pos {
	return l.start
}
func (l List) End() Pos {
	return l.end
}

func (l List) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T\n", l)

	for i, le := range l.Entries {
		if i+1 == len(l.Entries) {
			r += le.repr(style, pre2+style.Leaf, pre2+style.Blank)
		} else {
			r += le.repr(style, pre2+style.Branch, pre2+style.Carry)
		}
	}

	return r
}
