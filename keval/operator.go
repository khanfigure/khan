package keval

import (
	"reflect"

	"khan.rip/ksyn"
)

var opHandlers = map[ksyn.Operator][]reflect.Value{
	ksyn.OpCompare: []reflect.Value{
		reflect.ValueOf(func(a, b int) bool {
			return a == b
		}),
		reflect.ValueOf(func(a, b string) bool {
			return a == b
		}),
	},
	ksyn.OpNegCompare: []reflect.Value{
		reflect.ValueOf(func(a, b int) bool {
			return a != b
		}),
	},
	ksyn.OpAdd: []reflect.Value{
		reflect.ValueOf(func(a, b int) int {
			return a + b
		}),
		reflect.ValueOf(func(a, b string) string {
			return a + b
		}),
	},
	ksyn.OpMul: []reflect.Value{
		reflect.ValueOf(func(a, b int) int {
			return a * b
		}),
	},
	ksyn.OpDot: []reflect.Value{
		reflect.ValueOf(func(m map[string]interface{}, k string) interface{} {
			return m[k]
		}),
	},
}

func (m *Machine) eval_op(scope *Scope, node ksyn.Op, depth int) (Value, error) {
	handlers := opHandlers[node.Operator]

	if len(handlers) == 0 {
		return Value{}, ksyn.Errorf(node, "Unhandled operator %s", node.Operator)
	}

	if node.Operator.Unary() {
		if node.Operator.RightAssociative() {
			if node.Left != nil || node.Right == nil {
				return Value{}, ksyn.Errorf(node, "Unary right-associative operator requires one operand to the right")
			}
		} else {
			if node.Left == nil || node.Right != nil {
				return Value{}, ksyn.Errorf(node, "Unary left-associative operator requires one operand to the left")
			}
		}
	} else {
		if node.Left == nil || node.Right == nil {
			return Value{}, ksyn.Errorf(node, "Non-unary operator requires two operands")
		}
	}

	var (
		lv, rv   Value
		lvv, rvv reflect.Value
		err      error
	)

	if node.Left != nil {
		lv, err = m.eval(scope, node.Left, depth)
		if err != nil {
			return Value{}, err
		}
		lvv = reflect.ValueOf(lv.V)
	}

	if node.Right != nil {
		rv, err = m.eval(scope, node.Right, depth)
		if err != nil {
			return Value{}, err
		}
		rvv = reflect.ValueOf(rv.V)
	}

	for _, h := range handlers {
		t := h.Type()

		if t.NumOut() != 1 {
			continue
		}
		if node.Right == nil {
			// Unary left mode: Left hand operand is the only argument
			if t.NumIn() != 1 {
				continue
			}
			if rvv.Kind() != t.In(0).Kind() {
				continue
			}
			vals := h.Call([]reflect.Value{lvv})
			return Value{V: vals[0].Interface()}, nil
		} else if node.Left == nil {
			// Unary right mode: Right hand operand is the only argument
			if t.NumIn() != 1 {
				continue
			}
			if rvv.Kind() != t.In(0).Kind() {
				continue
			}
			vals := h.Call([]reflect.Value{rvv})
			return Value{V: vals[0].Interface()}, nil
		} else {
			// Regular mode: Left hand and right hand operands are passed as arguments
			if t.NumIn() != 2 {
				continue
			}
			if lvv.Kind() != t.In(0).Kind() {
				continue
			}
			if rvv.Kind() != t.In(1).Kind() {
				continue
			}
			vals := h.Call([]reflect.Value{lvv, rvv})
			return Value{V: vals[0].Interface()}, nil
		}
	}
	if node.Operator.Unary() {
		return Value{}, ksyn.Errorf(node, "Operator %s not valid for type %T", node.Operator, rv.V)
	}
	return Value{}, ksyn.Errorf(node, "Operator %s not valid between types %T and %T", node.Operator, lv.V, rv.V)
}
