package keval

import (
	"fmt"
	"reflect"

	"khan.rip/ksyn"

	"gopkg.in/yaml.v3"
	//"github.com/davecgh/go-spew/spew"
)

const MaxRecursion = 666

type Machine struct {
	scope *Scope
}

type Scope struct {
	symbols map[string]Value
	parent  *Scope
}

func (scope *Scope) Lookup(ident string) (Value, bool) {
	for scope != nil {
		v, ok := scope.symbols[ident]
		if ok {
			return v, true
		}
		scope = scope.parent
	}
	return Value{}, false
}

func builtin_itos(v int) string {
	return fmt.Sprintf("%d", v)
}
func builtin_println(s string) {
	fmt.Println(s)
}

// NewMachine() creates a new runtime but does not evaluate anything
func NewMachine() *Machine {
	m := Machine{}

	scope := &Scope{
		symbols: map[string]Value{},
	}

	scope.symbols["nil"] = Value{}

	scope.symbols["true"] = Value{
		V: true,
	}
	scope.symbols["false"] = Value{
		V: false,
	}

	scope.symbols["itos"] = Value{
		V: builtin_itos,
	}

	scope.symbols["println"] = Value{
		V: builtin_println,
	}

	scope.symbols["from_yaml"] = Value{
		V: builtin_from_yaml,
	}

	scope.symbols["include"] = Value{
		V: func(fpath string) {
			builtin_include(&m, fpath)
		},
	}

	m.scope = scope

	return &m
}
func (m *Machine) Eval(node ksyn.Node) (Value, error) {
	v, err := m.eval(m.scope, node, 0)
	if err != nil {
		return Value{}, err
	}

	return v, err
}

// Eval() creates a new runtime, executes the node, and returns the value
func Eval(node ksyn.Node) (Value, error) {
	m := NewMachine()
	v, err := m.Eval(node)
	return v, err
}

func (m *Machine) eval(scope *Scope, node ksyn.Node, depth int) (Value, error) {

	r := Value{}

	if nif, ok := node.(ksyn.IfStmt); ok {
		return m.eval_if(scope, nif, depth)
	}

	if nfor, ok := node.(ksyn.ForInStmt); ok {
		return m.eval_for_in(scope, nfor, depth)
	}

	if nblock, ok := node.(ksyn.BlockStmt); ok {

		newscope := &Scope{
			parent: scope,
		}

		var (
			lastval Value
			err     error
		)
		for _, stmt := range nblock.Stmts {
			lastval, err = m.eval(newscope, stmt, depth)
			if err != nil {
				return lastval, err
			}
		}
		return lastval, err
	}

	if nexprstmt, ok := node.(ksyn.ExprStmt); ok {
		return m.eval(scope, nexprstmt.Expr, depth)
	}

	if nstr, ok := node.(ksyn.Str); ok {
		return Value{V: nstr.Str}, nil
	}
	if nint, ok := node.(ksyn.Int); ok {
		return Value{V: nint.Int}, nil
	}

	/*if ncall, ok := node.(ksyn.CallExpr); ok {

		// resolve the function
		fv, err := m.eval(scope, *ncall.Ident, depth)
		if err != nil {
			return r, err
		}

		// evaluate arguments
		args := make([]Value, len(ncall.Args))
		for i, argnode := range ncall.Args {
			av, err := m.eval(scope, argnode, depth)
			if err != nil {
				return r, err
			}
			args[i] = av
		}

		return m.execute_function(ncall, fv, args, ncall.Args, scope, depth)
	}*/

	if nident, ok := node.(ksyn.Ident); ok {
		v, ok := scope.Lookup(nident.Ident)
		if !ok {
			return r, ksyn.Errorf(nident, "Unknown identifier %#v", nident.Ident)
		}
		return v, nil
	}

	if nfuncstmt, ok := node.(ksyn.FuncStmt); ok {
		return m.eval_func_decl(scope, nfuncstmt, depth)
	}

	if nfuncexpr, ok := node.(ksyn.FuncExpr); ok {
		return m.eval_func_expr(scope, nfuncexpr, depth)
	}

	if nop, ok := node.(ksyn.Op); ok {
		switch nop.Operator {
		case ksyn.OpAssign:
			return m.eval_assign(scope, nop, depth)
		case ksyn.OpDeclareAssign:
			return m.eval_decl_assign(scope, nop, depth)
		case ksyn.OpCall:
			fv, err := m.eval(scope, nop.Left, depth)
			if err != nil {
				return Value{}, err
			}

			ncall, ok := nop.Right.(ksyn.CallExpr)
			if !ok {
				return Value{}, ksyn.Errorf(nop.Right, "OpCall requires a CallExpr as the right operand")
			}

			// evaluate arguments
			args := make([]Value, len(ncall.Args))
			for i, argnode := range ncall.Args {
				av, err := m.eval(scope, argnode, depth)
				if err != nil {
					return Value{}, err
				}
				args[i] = av
			}
			return m.execute_function(ncall, fv, args, ncall.Args, scope, depth)
		default:
			return m.eval_op(scope, nop, depth)
		}
	}

	if nfile, ok := node.(ksyn.File); ok {
		var (
			lastval Value
			err     error
		)
		for _, stmt := range nfile.Stmts {
			lastval, err = m.eval(scope, stmt, depth)
			if err != nil {
				return lastval, err
			}
		}
		return lastval, err
	}

	return r, ksyn.Errorf(node, "Unhandled node type: %T", node)
}

func (m *Machine) eval_if(scope *Scope, node ksyn.IfStmt, depth int) (Value, error) {
	var (
		v   Value
		err error
	)

	for _, pair := range node.Pairs {
		v, err = m.eval(scope, pair.Expr, depth)
		if err != nil {
			return v, err
		}
		vb, ok := v.V.(bool)
		if !ok {
			return v, ksyn.Errorf(pair.Expr, "Cannot evaluate if block: Expected bool, got %#v", v.V)
		}
		if vb {
			return m.eval(scope, pair.Block, depth)
		}
	}

	// else?
	if node.Else != nil {
		return m.eval(scope, *node.Else, depth)
	}

	// i guess return the value of the most recently evaluated if expression?
	// this in theory should _always_ be false.
	return v, nil
}

func (m *Machine) eval_assign(scope *Scope, node ksyn.Op, depth int) (Value, error) {
	r := Value{}
	var err error

	if node.Left == nil || node.Right == nil {
		return r, ksyn.Errorf(node, "Assignment operator requires two operands")
	}

	ident, ok := node.Left.(ksyn.Ident)
	if !ok {
		return r, ksyn.Errorf(node, "Cannot assign into a %T", node.Left)
	}

	r, err = m.eval(scope, node.Right, depth)
	if err != nil {
		return r, err
	}

	for scope != nil {
		if scope.symbols == nil {
			scope = scope.parent
			continue
		}
		v, ok := scope.symbols[ident.Ident]
		if !ok {
			scope = scope.parent
			continue
		}
		if reflect.ValueOf(v.V).Kind() != reflect.ValueOf(r.V).Kind() {
			return r, ksyn.Errorf(ident, "Cannot assign %s to %s: Type mismatch (%T and %T)",
				r, ident.Ident, r.V, v.V)
		}
		scope.symbols[ident.Ident] = r
		return r, nil
	}

	return r, ksyn.Errorf(ident, "Identifier %#v not found: Cannot assign", ident.Ident)
}

func (m *Machine) eval_decl_assign(scope *Scope, node ksyn.Op, depth int) (Value, error) {
	r := Value{}
	var err error

	if node.Left == nil || node.Right == nil {
		return r, ksyn.Errorf(node, "Declare and assign operator requires two operands")
	}

	ident, ok := node.Left.(ksyn.Ident)
	if !ok {
		return r, ksyn.Errorf(node, "Cannot assign into a %T", node.Left)
	}

	if scope.symbols == nil {
		scope.symbols = make(map[string]Value)
	} else {
		_, ok := scope.symbols[ident.Ident]
		if ok {
			return r, ksyn.Errorf(ident, "Cannot redeclare %#v", ident.Ident)
		}
	}

	r, err = m.eval(scope, node.Right, depth)
	if err != nil {
		return r, err
	}

	scope.symbols[ident.Ident] = r
	return r, nil
}

func (m *Machine) execute_function(node ksyn.Node, fn Value, args []Value, argnodes []ksyn.Expr, scope *Scope, depth int) (Value, error) {

	// A khan function?
	nstmt, ok := fn.V.(ksyn.FuncStmt)
	if ok {
		return m.execute_function_node(scope, node, nstmt.Type, nstmt.Block, fn.Scope, args, argnodes, depth)
	}
	nexpr, ok := fn.V.(ksyn.FuncExpr)
	if ok {
		return m.execute_function_node(scope, node, nexpr.Type, nexpr.Block, fn.Scope, args, argnodes, depth)
	}

	// Otherwise assume a go function.

	fnv := reflect.ValueOf(fn.V)

	if fnv.Kind() != reflect.Func {
		return Value{}, ksyn.Errorf(node, "Cannot call: Type %T, expected function", fn.V)
	}

	fnt := fnv.Type()

	if len(args) != fnt.NumIn() {
		return Value{}, ksyn.Errorf(node, "Function argument count mismatch: Got %d, expected %d", len(args), fnt.NumIn())
	}

	if depth > MaxRecursion {
		return Value{}, ksyn.Errorf(node, "Cannot call: Hit maximum recursion depth (%d)", MaxRecursion)
	}

	go_args := make([]reflect.Value, len(args))
	for i, v := range args {
		rv := reflect.ValueOf(v.V)

		// Hack in some basic string coalescing?
		if fnt.In(i).Kind() == reflect.String {
			switch rvv := v.V.(type) {
			case yaml.Node:
				if rvv.Kind == yaml.ScalarNode {
					rv = reflect.ValueOf(rvv.Value)
				}
			case *yaml.Node:
				if rvv == nil {
					rv = reflect.ValueOf(nil)
				} else {
					if rvv.Kind == yaml.ScalarNode {
						rv = reflect.ValueOf(rvv.Value)
					}
				}
			}
		}

		if rv.Kind() != fnt.In(i).Kind() {

			return Value{}, ksyn.Errorf(argnodes[i], "Function argument %d type mismatch: Got %s, expected %s", i, rv.Kind(), fnt.In(i).Kind())
		}

		go_args[i] = rv
	}

	go_ret := fnv.Call(go_args)

	if len(go_ret) == 0 {
		return Value{}, nil
	}

	if len(go_ret) == 1 {
		rr := Value{V: go_ret[0].Interface()}
		return rr, nil
	}

	rv := make([]interface{}, len(go_ret))
	for i, v := range go_ret {
		rv[i] = v.Interface()
	}

	return Value{V: rv}, nil
}

func (m *Machine) execute_function_node(scope *Scope, node ksyn.Node, ftype ksyn.FuncType, body ksyn.BlockStmt, fscope *Scope, args []Value, argnodes []ksyn.Expr, depth int) (Value, error) {

	if fscope != nil {
		scope = fscope
	}

	if len(args) != len(ftype.Args) {
		return Value{}, ksyn.Errorf(node, "Function argument count mismatch: Got %d, expected %d", len(args), len(ftype.Args))
	}

	// if there are variable declarations, wrap scope
	if len(args) > 0 {
		scope = &Scope{
			parent:  scope,
			symbols: make(map[string]Value, len(args)),
		}
		for i, v := range args {
			n := ftype.Args[i].Ident.Ident
			// TODO type check with function argument type.
			scope.symbols[n] = v
		}
	}

	return m.eval(scope, body, depth+1)
}

func (m *Machine) eval_func_decl(scope *Scope, node ksyn.FuncStmt, depth int) (Value, error) {
	r := Value{}

	if scope.symbols == nil {
		scope.symbols = make(map[string]Value)
	} else {
		_, ok := scope.symbols[node.Ident.Ident]
		if ok {
			return r, ksyn.Errorf(node.Ident, "Cannot redeclare %#v", node.Ident.Ident)
		}
	}

	r.V = node
	r.Scope = scope

	scope.symbols[node.Ident.Ident] = r

	// A named function declaration doesn't have a return value, I guess?
	return Value{}, nil
}

func (m *Machine) eval_func_expr(scope *Scope, node ksyn.FuncExpr, depth int) (Value, error) {
	r := Value{}
	r.V = node
	return r, nil
}
