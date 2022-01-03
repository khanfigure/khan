package ksyn

import (
	"fmt"
)

type BlockStmt struct {
	Stmts []Stmt

	start, end Pos
}

type BlockExpr struct {
	Expr       Expr
	start, end Pos
}

func (b BlockStmt) Start() Pos {
	return b.start
}

func (b BlockStmt) End() Pos {
	return b.end
}

func (p *parser) block() (BlockStmt, error) {
	start := p.pos

	if err := p.consumeRune('{'); err != nil {
		return BlockStmt{}, err
	}

	p.chomp()

	if _, err := p.comments(); err != nil {
		return BlockStmt{}, err
	}

	stmts, err := p.stmts()
	if err != nil {
		return BlockStmt{}, err
	}

	if _, err := p.comments(); err != nil {
		return BlockStmt{}, err
	}

	if err := p.consumeRune('}'); err != nil {
		return BlockStmt{}, err
	}

	end := p.pos

	return BlockStmt{
		Stmts: stmts,

		start: start,
		end:   end,
	}, nil
}

func (b BlockStmt) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T\n", b)
	r += reprStmts(b.Stmts, style, pre1, pre2)
	return r
}

func (b BlockExpr) Start() Pos {
	return b.start
}

func (b BlockExpr) End() Pos {
	return b.end
}

func (p *parser) blockExpr() (BlockExpr, error) {
	start := p.pos

	if err := p.consumeRune('{'); err != nil {
		return BlockExpr{}, err
	}

	p.chomp()

	if _, err := p.comments(); err != nil {
		return BlockExpr{}, err
	}

	expr, err := p.expr()
	if err != nil {
		return BlockExpr{}, err
	}

	if _, err := p.comments(); err != nil {
		return BlockExpr{}, err
	}

	p.chomp()

	if err := p.consumeRune('}'); err != nil {
		return BlockExpr{}, err
	}

	end := p.pos

	return BlockExpr{
		Expr: expr,

		start: start,
		end:   end,
	}, nil
}

func (b BlockExpr) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T\n", b)
	r += b.Expr.repr(style, pre2+style.Leaf, pre2+style.Blank)
	return r
}
