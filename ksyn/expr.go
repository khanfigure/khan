package ksyn

import (
	"fmt"
)

type Expr interface {
	Node
	IsExpr()
}

// All possible expressions
func (_ Op) IsExpr()        {}
func (_ Str) IsExpr()       {}
func (_ Num) IsExpr()       {}
func (_ RGB) IsExpr()       {}
func (_ Dict) IsExpr()      {}
func (_ List) IsExpr()      {}
func (_ Map) IsExpr()       {}
func (_ FuncExpr) IsExpr()  {}
func (_ CallExpr) IsExpr()  {}
func (_ Select) IsExpr()    {}
func (_ Ident) IsExpr()     {}
func (_ BlockExpr) IsExpr() {}

// Type expressions
func (_ FuncType) IsExpr() {}

// Wikipedia was a big help for getting this right.
// My first few stabs were super wrong.
// https://en.wikipedia.org/wiki/Operator-precedence_parser
func (p *parser) expr() (Expr, error) {
	if _, err := p.comments(); err != nil {
		return nil, err
	}

	lhs, err := p.exprPrimary()
	if err != nil {
		return nil, err
	}

	if _, err := p.comments(); err != nil {
		return nil, err
	}

	v, err := p.exprOp(lhs, 0)
	if err != nil {
		return nil, err
	}

	if _, err := p.comments(); err != nil {
		return nil, err
	}

	return v, nil
}

func (p *parser) exprPrimary() (Expr, error) {
	p.chomp()
	ch := p.ch()
	str := p.str()

	if ch_is_num(ch) {
		return p.num()
	}
	if ch == '"' {
		return p.exprStr()
	}
	if ch == '#' {
		return p.rgb()
	}
	if hasPrefixToken(str, "map") {
		return p.parseMap()
	}
	if hasPrefixToken(str, "func") {
		return p.parseFuncExpr()
	}
	if hasPrefixFold(str, "select") {
		return p.parseSelect()
	}
	if ch == '{' {
		return p.exprDict()
	}
	if ch == '(' {
		return p.exprParen()
	}
	if ch == '!' {
		// unary operator: return an empty left hand side
		return nil, nil
	}
	if ch_is_ident(ch) {
		// looks like an identifier
		ident, err := p.ident()
		if err != nil {
			return nil, err
		}
		p.chomp()
		return ident, nil
	}
	return nil, p.unexpected("an expression")
}

func (p *parser) exprOp(lhs Expr, min_prec int) (Expr, error) {
	lookahead := p.strPeekChompComments()

	for {
		op, opl := opFromString(lookahead)
		if op == OpInvalid {
			break
		}
		if op.precedence() < min_prec {
			break
		}

		// consume comments and whitespace before the operator.
		p.chomp()
		if _, err := p.comments(); err != nil {
			return nil, err
		}
		p.chomp()

		oppos := p.pos
		p.move(opl)
		opposend := p.pos

		// consume comments after the operator
		if _, err := p.comments(); err != nil {
			return nil, err
		}

		var (
			rhs Expr
			err error
		)

		if op == OpCall {
			args, err := p.exprParenList()
			if err != nil {
				return nil, err
			}
			rhs = CallExpr{
				Args:  args,
				start: oppos,
				end:   p.pos,
			}
		} else {
			rhs, err = p.exprPrimary()
			if err != nil {
				return nil, err
			}
		}

		lookahead = p.strPeekChompComments()
		for {
			op2, _ := opFromString(lookahead)
			if op2 == OpInvalid {
				break
			}

			p1 := op.precedence()
			p2 := op2.precedence()

			if p2 > p1 || (p2 == p1 && op2.RightAssociative()) {
				rhs, err = p.exprOp(rhs, p2)
				if err != nil {
					return nil, err
				}
				lookahead = p.strPeekChompComments()
			} else {
				break
			}
		}

		lhsReplace := Op{
			Operator: op,

			start: oppos,
			end:   opposend,
		}

		// in the case of a unary operator, there shouldn't be a left hand side
		if lhs == nil {
			if !op.Unary() {
				return nil, Error{
					Err: fmt.Errorf("Operator with only right hand side must be unary: %v", op),
					Pos: oppos,
				}
			}
			// cool
		} else {
			lhsReplace.Left = lhs

			// Expand the position of the oprator to include its left operand
			lhsReplace.start = lhs.Start()
		}

		// Expand the position of the operator to inlcude its right operand
		lhsReplace.Right = rhs
		lhsReplace.end = rhs.End()

		lhs = lhsReplace
	}

	return lhs, nil
}

func (p *parser) exprParen() (Expr, error) {
	if err := p.consumeRune('('); err != nil {
		return nil, err
	}

	r, err := p.expr()
	if err != nil {
		return nil, err
	}

	p.chomp()

	if err := p.consumeRune(')'); err != nil {
		return nil, err
	}

	return r, nil
}

func (p *parser) exprParenList() ([]Expr, error) {

	p.chomp()
	ch := p.ch()
	if ch == ')' {
		p.next()
		return nil, nil
	}

	var r []Expr

	for {
		p.chomp()

		expr, err := p.expr()
		if err != nil {
			return nil, err
		}
		r = append(r, expr)

		p.chomp()
		if p.ch() == ',' {
			p.next()
			continue
		}

		if err := p.consumeRune(')'); err != nil {
			return nil, err
		}

		return r, nil
	}
}

/*func reprExprs(exprs []Expr, style reprStyle, pre1, pre2 string) string {
	r := ""

	for i, c := range exprs {
		c1 := style.Branch
		c2 := style.Carry
		if i + 1 == len(exprs) {
			c1 = style.Leaf
			c2 = style.Blank
		}
		r += c.repr(style, pre2 + c1, pre2 + c2)
	}

	return r
}
*/
