package ksyn

import (
	"fmt"
)

type Map struct {
	Expr  Expr
	Block BlockStmt

	start, end Pos
}

func (m Map) Start() Pos {
	return m.start
}
func (m Map) End() Pos {
	return m.end
}

func (p *parser) parseMap() (Map, error) {
	start := p.pos

	if err := p.consume("map"); err != nil {
		return Map{}, err
	}

	expr, err := p.expr()
	if err != nil {
		return Map{}, err
	}

	p.chomp()

	block, err := p.block()
	if err != nil {
		return Map{}, err
	}

	end := p.pos

	return Map{
		Expr:  expr,
		Block: block,

		start: start,
		end:   end,
	}, nil
}

func (m Map) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T\n", m)

	r += m.Expr.repr(style, pre2+style.Branch, pre2+style.Carry)
	r += m.Block.repr(style, pre2+style.Leaf, pre2+style.Blank)

	return r
}
