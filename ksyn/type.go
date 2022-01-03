package ksyn

import (
	"fmt"
)

type FuncType struct {
	Args    []Var
	Returns []Var

	start, end Pos
}

func (f FuncType) Start() Pos {
	return f.start
}
func (f FuncType) End() Pos {
	return f.end
}

func (p *parser) parseFuncType() (FuncType, error) {
	start := p.pos
	if err := p.consume("func"); err != nil {
		return FuncType{}, err
	}
	p.chomp()

	args, err := p.varsParen(varSpec{ArgMode: true})
	if err != nil {
		return FuncType{}, err
	}

	p.space() // newline not allowed here. can't remember why

	var returns []Var

	if p.ch() == '(' {
		returns, err = p.varsParen(varSpec{ReturnMode: true})
		if err != nil {
			return FuncType{}, err
		}
	} else if p.peekTypeExpr() {
		singleReturn, err := p.parseVar(varSpec{ReturnMode: true})
		if err != nil {
			return FuncType{}, err
		}
		returns = append(returns, singleReturn)
	} else {
		// no return value
	}

	end := p.pos

	return FuncType{
		Args:    args,
		Returns: returns,

		start: start,
		end:   end,
	}, nil
}

func (f FuncType) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T\n", f)

	for i, v := range f.Args {
		if i+1 == len(f.Args)+len(f.Returns) {
			r += repr_title("arg", v, style, pre2+style.Leaf, pre2+style.Blank)
		} else {
			r += repr_title("arg", v, style, pre2+style.Branch, pre2+style.Carry)
		}
	}
	for i, v := range f.Returns {
		if i+1 == len(f.Returns) {
			r += repr_title("return", v, style, pre2+style.Leaf, pre2+style.Blank)
		} else {
			r += repr_title("return", v, style, pre2+style.Branch, pre2+style.Carry)
		}
	}

	return r
}

func (p *parser) peekTypeExpr() bool {
	str := p.strPeekChompComments()

	if peekIdentStr(str) {
		return true
	}
	if hasPrefixToken(str, "func") {
		return true
	}

	return false
}

func (p *parser) typeExpr() (Expr, error) {
	p.chomp()
	if _, err := p.comments(); err != nil {
		return nil, err
	}
	p.chomp()

	str := p.str()

	if hasPrefixToken(str, "func") {
		return p.parseFuncType()
	}

	if p.peekIdent() {
		return p.ident()
	}

	return nil, p.unexpected("an identifier or type expression")
}
