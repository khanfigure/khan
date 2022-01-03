package ksyn

import (
	"fmt"
	"strings"
)

type Op struct {
	Operator Operator
	Left     Expr
	Right    Expr

	start, end Pos
}

func (o Op) Start() Pos {
	return o.start
}
func (o Op) End() Pos {
	return o.end
}

type Operator int

const (
	OpNone Operator = iota
	OpInvalid
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpAssign
	OpDeclareAssign
	OpDeclareAssignReactive
	OpCompare
	OpNegCompare
	OpNot
	OpDot  // select fields out of structures
	OpCall // call expression
)

func (o Operator) String() string {
	switch o {
	case OpNone:
		return "OpNone"
	case OpAdd:
		return "OpAdd"
	case OpSub:
		return "OpSub"
	case OpMul:
		return "OpMul"
	case OpDiv:
		return "OpDiv"
	case OpAssign:
		return "OpAssign"
	case OpDeclareAssign:
		return "OpDeclareAssign"
	case OpDeclareAssignReactive:
		return "OpDeclareAssignReactive"
	case OpCompare:
		return "OpCompare"
	case OpNegCompare:
		return "OpNegCompare"
	case OpNot:
		return "OpNot"
	case OpDot:
		return "OpDot"
	case OpCall:
		return "OpCall"
	default:
		return fmt.Sprintf("OpInvalid(%d)", o)
	}
}

// returns Operator and a length of the consumed source string
func opFromString(s string) (Operator, int) {
	if len(s) == 0 {
		return OpInvalid, 0
	}

	// Make sure we can have <tag key=value/> where the expression is "value"
	// and / is not treated as a divide operator.
	if strings.HasPrefix(s, "/>") {
		return OpInvalid, 0
	}

	if strings.HasPrefix(s, ":=") {
		return OpDeclareAssign, 2
	}
	if strings.HasPrefix(s, "~=") {
		return OpDeclareAssignReactive, 2
	}
	if strings.HasPrefix(s, "==") {
		return OpCompare, 2
	}
	if strings.HasPrefix(s, "!=") {
		return OpNegCompare, 2
	}
	switch s[0] {
	case '+':
		return OpAdd, 1
	case '-':
		return OpSub, 1
	case '*':
		return OpMul, 1
	case '/':
		return OpDiv, 1
	case '=':
		return OpAssign, 1
	case '!':
		return OpNot, 1
	case '.':
		return OpDot, 1
	case '(':
		return OpCall, 1
	}
	return OpInvalid, 0
}

func (o Operator) precedence() int {
	switch o {
	case OpAdd:
		return 1
	case OpSub:
		return 1
	case OpMul:
		return 2
	case OpDiv:
		return 2
	case OpNot:
		return 3
	case OpCall:
		return 4
	}

	// Don't return anything < 0 unless you
	// fix the initial value in parse.parseExpr
	return 0
}

func (o Operator) RightAssociative() bool {
	switch o {
	case OpNot:
		return true
	}
	return false
}

func (o Operator) Unary() bool {
	switch o {
	case OpNot:
		return true
	}
	return false
}

func (o Op) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1 + fmt.Sprintf("%T %s\n", o, o.Operator)

	var children []Node

	if o.Left != nil {
		children = append(children, o.Left)
	}
	if o.Right != nil {
		children = append(children, o.Right)
	}

	for i, c := range children {
		c1 := style.Branch
		c2 := style.Carry
		if i+1 == len(children) {
			c1 = style.Leaf
			c2 = style.Blank
		}
		r += c.repr(style, pre2+c1, pre2+c2)
	}

	return r
}
