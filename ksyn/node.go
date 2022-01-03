package ksyn

type Node interface {
	IsNode()

	Start() Pos
	End() Pos

	repr(reprStyle, string, string) string
}

// All possible nodes
func (_ File) IsNode() {}

// Statements
func (_ BlockStmt) IsNode() {}
func (_ IfStmt) IsNode()    {}
func (_ FuncStmt) IsNode()  {}
func (_ ExprStmt) IsNode()  {}
func (_ VarStmt) IsNode()   {}
func (_ ForInStmt) IsNode() {}
func (_ EmitStmt) IsNode()  {}

// Expressions
func (_ Op) IsNode()        {}
func (_ Str) IsNode()       {}
func (_ Int) IsNode()       {}
func (_ Dict) IsNode()      {}
func (_ RGB) IsNode()       {}
func (_ Map) IsNode()       {}
func (_ FuncExpr) IsNode()  {}
func (_ CallExpr) IsNode()  {}
func (_ Select) IsNode()    {}
func (_ Ident) IsNode()     {}
func (_ BlockExpr) IsNode() {}

// Type Expressions
func (_ FuncType) IsNode() {}

// Danglings
func (_ SelectExpr) IsNode() {}
func (_ SelectFrom) IsNode() {}
func (_ Var) IsNode()        {}
