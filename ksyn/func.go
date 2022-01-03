package ksyn

import (
	"fmt"
)

type CallExpr struct {
	Args []Expr

	start, end Pos
}

func (c CallExpr) Start() Pos {
	return c.start
}
func (c CallExpr) End() Pos {
	return c.end
}
func (c CallExpr) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T\n", c)

	for i, a := range c.Args {
		if i+1 == len(c.Args) {
			r += a.repr(style, pre2+style.Leaf, pre2+style.Blank)
		} else {
			r += a.repr(style, pre2+style.Branch, pre2+style.Carry)
		}
	}

	return r
}

type FuncStmt struct {
	Ident Ident
	Type  FuncType
	Block BlockStmt

	start, end Pos
}

func (f FuncStmt) Start() Pos {
	return f.start
}
func (f FuncStmt) End() Pos {
	return f.end
}

type FuncExpr struct {
	Type  FuncType
	Block BlockStmt

	start, end Pos
}

func (f FuncExpr) Start() Pos {
	return f.start
}
func (f FuncExpr) End() Pos {
	return f.end
}

func (p *parser) parseFunc() (FuncStmt, error) {
	start := p.pos

	// Since a function declaration has an identifier in the middle,
	// we don't share the code with parseFuncType. Could be merged
	// with some work though.

	if err := p.consume("func"); err != nil {
		return FuncStmt{}, err
	}
	p.chomp()

	ident, err := p.ident()
	if err != nil {
		return FuncStmt{}, err
	}

	p.chomp()

	args, err := p.varsParen(varSpec{ArgMode: true})
	if err != nil {
		return FuncStmt{}, err
	}

	p.space() // newline not allowed here. can't remember why

	var returns []Var

	if p.ch() == '(' {
		returns, err = p.varsParen(varSpec{ReturnMode: true})
		if err != nil {
			return FuncStmt{}, err
		}
	} else if p.peekTypeExpr() {
		singleReturn, err := p.parseVar(varSpec{ReturnMode: true})
		if err != nil {
			return FuncStmt{}, err
		}
		returns = append(returns, singleReturn)
	} else {
		// no return value
	}

	typeend := p.pos

	p.space() // i guess newline not allowed here either?

	// function body
	stmt, err := p.block()
	if err != nil {
		return FuncStmt{}, err
	}

	end := p.pos

	// Simulate a parsed function type
	ft := FuncType{
		Args:    args,
		Returns: returns,
		start:   start,
		end:     typeend,
	}

	return FuncStmt{
		Ident: ident,
		Type:  ft,
		Block: stmt,

		start: start,
		end:   end,
	}, nil
}

func (p *parser) parseFuncExpr() (FuncExpr, error) {
	start := p.pos

	ft, err := p.parseFuncType()
	if err != nil {
		return FuncExpr{}, err
	}

	p.space() // i guess newline not allowed here either?

	// function body
	stmt, err := p.block()
	if err != nil {
		return FuncExpr{}, err
	}

	end := p.pos

	return FuncExpr{
		Type:  ft,
		Block: stmt,

		start: start,
		end:   end,
	}, nil
}

func (f FuncStmt) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T\n", f)

	r += f.Ident.repr(style, pre2+style.Branch, pre2+style.Carry)
	r += f.Type.repr(style, pre2+style.Branch, pre2+style.Carry)
	r += f.Block.repr(style, pre2+style.Leaf, pre2+style.Blank)

	return r
}

func (f FuncExpr) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T\n", f)

	r += f.Type.repr(style, pre2+style.Branch, pre2+style.Carry)
	r += f.Block.repr(style, pre2+style.Leaf, pre2+style.Blank)

	return r
}
