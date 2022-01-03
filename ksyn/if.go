package ksyn

import (
	"fmt"
)

type IfStmt struct {
	Pairs []IfPair
	Else  *BlockStmt

	start, end Pos
}

type IfPair struct {
	Expr  Expr
	Block BlockStmt
}

func (i IfStmt) Start() Pos {
	return i.start
}
func (i IfStmt) End() Pos {
	return i.end
}

func (p *parser) iff() (IfStmt, error) {
	r := IfStmt{
		start: p.pos,
	}

	for i := 0; ; i++ {
		if err := p.consume("if"); err != nil {
			return r, err
		}

		expr, err := p.expr()
		if err != nil {
			return r, err
		}

		p.chomp()

		block, err := p.block()
		if err != nil {
			return r, err
		}

		r.Pairs = append(r.Pairs, IfPair{
			Expr:  expr,
			Block: block,
		})

		p.chomp()

		if hasPrefixToken(p.str(), "else") {
			p.consume("else")
			p.chomp()
			if hasPrefixToken(p.str(), "if") {
				continue
			}
			final, err := p.block()
			if err != nil {
				return r, err
			}
			r.Else = &final
		}

		break
	}

	r.end = p.pos

	return r, nil
}

func (i IfStmt) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1 + fmt.Sprintf("%T\n", i)

	var children []Node

	for _, c := range i.Pairs {
		children = append(children, c.Expr)
		children = append(children, c.Block)
	}
	if i.Else != nil {
		children = append(children, *i.Else)
	}

	for ii, c := range children {
		if ii+1 == len(children) {
			r += c.repr(style, pre2+style.Leaf, pre2+style.Blank)
		} else {
			r += c.repr(style, pre2+style.Branch, pre2+style.Carry)
		}
	}
	return r
}
