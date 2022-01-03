package ksyn

import (
	"fmt"
)

type ForInStmt struct {
	Ident      Ident
	Expr       Expr
	Block      BlockStmt
	start, end Pos
}

func (fi ForInStmt) Start() Pos {
	return fi.start
}

func (fi ForInStmt) End() Pos {
	return fi.end
}

func (fi ForInStmt) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1 + fmt.Sprintf("%T\n", fi)
	r += fi.Ident.repr(style, pre2+style.Branch, pre2+style.Carry)
	r += fi.Expr.repr(style, pre2+style.Branch, pre2+style.Carry)
	r += fi.Block.repr(style, pre2+style.Leaf, pre2+style.Blank)
	return r
}

func (p *parser) parseForInStmt() (ForInStmt, error) {
	r := ForInStmt{}
	r.start = p.pos

	if err := p.consume("for"); err != nil {
		return r, err
	}

	p.chomp()

	ident, err := p.ident()
	if err != nil {
		return r, err
	}

	p.chomp()

	if err := p.consume("in"); err != nil {
		return r, err
	}

	p.chomp()

	expr, err := p.expr()
	if err != nil {
		return r, err
	}

	r.end = p.pos

	p.chomp()

	stmt, err := p.block()
	if err != nil {
		return r, err
	}

	r.Ident = ident
	r.Expr = expr
	r.Block = stmt

	return r, nil
}
