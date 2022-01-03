package ksyn

import ()

type Stmt interface {
	Node
	IsStmt()
}

// All possible statements
func (_ BlockStmt) IsStmt() {}
func (_ IfStmt) IsStmt()    {}
func (_ FuncStmt) IsStmt()  {}
func (_ ExprStmt) IsStmt()  {}
func (_ VarStmt) IsStmt()   {}
func (_ ForInStmt) IsStmt() {}
func (_ EmitStmt) IsStmt()  {}

func (p *parser) stmts() ([]Stmt, error) {
	var r []Stmt

	for {
		ch := p.ch()
		if ch == 0 || ch == '}' {
			break
		}
		if ch == ' ' || ch == '\t' {
			p.move(1)
			continue
		}
		if ch == '\n' || ch == ';' {
			p.next()
			continue
		}

		_, err := p.comments()
		if err != nil {
			return nil, err
		}

		str := p.str()

		var (
			child Stmt
		)

		if ch == '{' {
			child, err = p.block()
		} else if hasPrefixToken(str, "if") {
			child, err = p.iff()
		} else if hasPrefixToken(str, "func") {
			child, err = p.parseFunc()
		} else if hasPrefixToken(str, "var") {
			child, err = p.parseVarStmt()
		} else if hasPrefixToken(str, "for") {
			child, err = p.parseForInStmt()
		} else if hasPrefixToken(str, "emit") {
			child, err = p.parseEmitStmt()
		} else {
			child, err = p.exprStmt()
		}
		if err != nil {
			return nil, err
		}

		if _, err := p.comments(); err != nil {
			return nil, err
		}

		r = append(r, child)
	}

	return r, nil
}

func reprStmts(stmts []Stmt, style reprStyle, pre1, pre2 string) string {
	r := ""

	for i, c := range stmts {
		c1 := style.Branch
		c2 := style.Carry
		if i+1 == len(stmts) {
			c1 = style.Leaf
			c2 = style.Blank
		}
		r += c.repr(style, pre2+c1, pre2+c2)
	}

	return r
}
