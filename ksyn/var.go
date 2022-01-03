package ksyn

import (
	"fmt"
)

type VarStmt struct {
	Vars []Var

	start, end Pos
}

type Var struct {
	Ident Ident
	Type  Expr
	Expr  Expr

	start, end Pos
}

type varSpec struct {
	VarMode    bool
	ArgMode    bool
	ReturnMode bool
}

// type lard int
// type (
//   lard int
// )
// var a = 5
// var (
//   a, b lard                         X not supported for now
//   c lard
//   d lard = 6
// )
// func a(b lard) lard {
// }
// var a func(lard) lard
// var a = func(b lard) lard
// func a(b = 5, c lard = 6, d lard)   X not supported
//

func (v VarStmt) Start() Pos {
	return v.start
}
func (v VarStmt) End() Pos {
	return v.end
}
func (v Var) Start() Pos {
	return v.start
}
func (v Var) End() Pos {
	return v.end
}
func (v VarStmt) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1 + fmt.Sprintf("%T\n", v)
	for i, vv := range v.Vars {
		if i+1 == len(v.Vars) {
			r += vv.repr(style, pre2+style.Leaf, pre2+style.Blank)
		} else {
			r += vv.repr(style, pre2+style.Branch, pre2+style.Carry)
		}
	}
	return r
}
func (v Var) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T\n", v)

	var children []Node
	var titles []string

	if v.Ident.Ident != "" {
		children = append(children, v.Ident)
		titles = append(titles, "ident")
	}

	if v.Type != nil {
		children = append(children, v.Type)
		titles = append(titles, "type")
	}

	if v.Expr != nil {
		children = append(children, v.Expr)
		titles = append(titles, "expr")
	}

	for i, c := range children {
		if i+1 == len(children) {
			r += repr_title(titles[i], c, style, pre2+style.Leaf, pre2+style.Blank)
		} else {
			r += repr_title(titles[i], c, style, pre2+style.Branch, pre2+style.Carry)
		}
	}
	return r
}

func (p *parser) parseVarStmt() (VarStmt, error) {
	start := p.pos
	if err := p.consume("var"); err != nil {
		return VarStmt{}, err
	}
	p.chomp()

	var (
		vars []Var
	)

	spec := varSpec{
		VarMode: true,
	}

	if p.ch() == '(' {
		// Set of many
		var err error
		vars, err = p.varsParen(spec)
		if err != nil {
			return VarStmt{}, err
		}
	} else {
		// Just one var
		vn, err := p.parseVar(spec)
		if err != nil {
			return VarStmt{}, err
		}
		vars = []Var{vn}
	}
	end := p.pos
	return VarStmt{
		Vars:  vars,
		start: start,
		end:   end,
	}, nil
}

func (p *parser) parseVar(spec varSpec) (Var, error) {
	start := p.pos

	vn := Var{
		start: start,
	}

	var (
		a, b Expr
		err  error
	)

	a, err = p.typeExpr()
	if err != nil {
		return Var{}, err
	}

	more := false
	if !spec.ReturnMode {
		more = p.peekTypeExpr()
	}

	if more {
		b, err = p.typeExpr()
		if err != nil {
			return Var{}, err
		}
		id, ok := a.(Ident)
		if !ok {
			return Var{}, NewErrFromNode(a, fmt.Errorf("Identifier expected: Got %T", a))
		}

		vn.Ident = id
		vn.Type = b
	} else {
		if spec.VarMode {
			id, ok := a.(Ident)
			if !ok {
				return Var{}, NewErrFromNode(a, fmt.Errorf("Identifier expected: Got %T", a))
			}
			vn.Ident = id
		} else {
			vn.Type = a
		}
	}

	vn.end = p.pos

	if spec.ReturnMode {
		return vn, nil
	}

	p.chomp()
	if _, err := p.comments(); err != nil {
		return Var{}, err
	}
	p.chomp()

	op, opl := opFromString(p.str())
	if op == OpAssign {
		p.move(opl)

		expr, err := p.expr()
		if err != nil {
			return Var{}, err
		}

		vn.Expr = expr
		vn.end = p.pos
	} else if spec.VarMode && vn.Type == nil {
		return Var{}, NewErrFromNode(a, fmt.Errorf("Type or assignment expression is required"))
	}

	if _, err := p.comments(); err != nil {
		return Var{}, err
	}
	p.chomp()

	return vn, nil
}

func (p *parser) varsParen(spec varSpec) ([]Var, error) {
	p.chomp()

	if err := p.consumeRune('('); err != nil {
		return nil, err
	}

	var r []Var

	for {
		p.chomp()

		ch := p.ch()
		if ch == ')' {
			p.next()
			return r, nil
		}

		// look for one or two identifiers separated by space,
		// and then for more variables after commas

		vn, err := p.parseVar(spec)
		if err != nil {
			return nil, err
		}

		r = append(r, vn)

		p.chomp()
		ch = p.ch()
		if ch == ';' || ch == ',' {
			// read another
			p.next()
			continue
		}

		// EXPERIMENTAL
		// read another even with no separator
		//		if p.peekIdent() {
		//			continue
		//		}
		//		if err := p.consumeRune(')'); err != nil {
		//			return nil, err
		//		}
		//		return r, nil
	}
}
