package ksyn

import (
	"fmt"
)

type EmitStmt struct {
	Expr       Expr
	start, end Pos
}

func (e EmitStmt) Start() Pos {
	return e.start
}
func (e EmitStmt) End() Pos {
	return e.end
}

func (e EmitStmt) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1 + fmt.Sprintf("%T\n", e)
	r += e.Expr.repr(style, pre2+style.Leaf, pre2+style.Blank)
	return r
}

func (p *parser) parseEmitStmt() (EmitStmt, error) {
	r := EmitStmt{start: p.pos}

	if err := p.consume("emit"); err != nil {
		return r, err
	}

	p.chomp()

	expr, err := p.expr()
	if err != nil {
		return r, err
	}

	r.Expr = expr
	r.end = p.pos
	return r, nil
}
