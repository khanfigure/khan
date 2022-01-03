package ksyn

import (
	"fmt"
)

type ExprStmt struct {
	Expr       Expr
	start, end Pos
}

func (es ExprStmt) Start() Pos {
	return es.start
}

func (es ExprStmt) End() Pos {
	return es.end
}

func (p *parser) exprStmt() (ExprStmt, error) {
	start := p.pos
	expr, err := p.expr()
	if err != nil {
		return ExprStmt{}, err
	}
	end := p.pos
	return ExprStmt{
		Expr:  expr,
		start: start,
		end:   end,
	}, nil
}

func (es ExprStmt) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T\n", es)
	r += es.Expr.repr(style, pre2+style.Leaf, pre2+style.Blank)
	return r
}
