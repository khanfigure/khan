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

/*func (m *Machine) eval_compare(scope *Scope, node ksyn.Op, neg bool, depth int) (Value, error) {
	var (
		r   Value
		err error
	)

	if node.Left == nil || node.Right == nil {
		return r, fmt.Errorf("Compare operator requires two operands")
	}
	v1, err := m.eval(scope, node.Left, depth)
	if err != nil {
		return r, err
	}
	v2, err := m.eval(scope, node.Right, depth)
	if err != nil {
		return r, nil
	}

	c := false

	switch a := v1.V.(type) {
	case bool:
		b, ok := v2.V.(bool)
		if !ok {
			return r, ksyn.Errorf(node, "Cannot compare types %T and %T", v1.V, v2.V)
		}
		c = (a == b)
	case string:
		b, ok := v2.V.(string)
		if !ok {
			return r, ksyn.Errorf(node, "Cannot compare types %T and %T", v1.V, v2.V)
		}
		c = (a == b)
	case int:
		b, ok := v2.V.(int)
		if !ok {
			return r, ksyn.Errorf(node, "Cannot compare types %T and %T", v1.V, v2.V)
		}
		c = (a == b)
	default:
		return r, ksyn.Errorf(node.Left, "Type %T is not comparable", v1.V)
	}

	if neg {
		c = !c
	}

	return Value{V: c}, nil
}

func (m *Machine) eval_not_op(scope *Scope, node ksyn.Op, depth int) (Value, error) {
	var (
		r   Value
		err error
	)

	if node.Left == nil || node.Right != nil {
		return r, ksyn.Errorf(node, "Not operator requires single operand")
	}
	v, err := m.eval(scope, node.Left, depth)
	if err != nil {
		return r, err
	}
	vb, ok := v.V.(bool)
	if !ok {
		return r, ksyn.Errorf(node.Left, "Not operator requires boolean value: Got %T", v.V)
	}
	return Value{V: vb}, nil
}

func (m *Machine) eval_int_op(scope *Scope, node ksyn.Op, fn func(a, b int) int, depth int) (Value, error) {
	var r int

	if node.Left == nil || node.Right == nil {
		return Value{}, ksyn.Errorf(node, "Operator requires two operands")
	}

	lv, err := m.eval(scope, node.Left, depth)
	if err != nil {
		return Value{}, err
	}
	rv, err := m.eval(scope, node.Right, depth)
	if err != nil {
		return Value{}, err
	}

	lvi, ok := lv.V.(int)
	if !ok {
		return Value{}, ksyn.Errorf(node.Left, "Expected int: Got %T", lv.V)
	}

	rvi, ok := rv.V.(int)
	if !ok {
		return Value{}, ksyn.Errorf(node.Right, "Expected int: Got %T", rv.V)
	}

	r = fn(lvi, rvi)

	return Value{V: r}, nil
}

type typefnmap struct {
	intintint func(int, int) int
	strstrstr func(string, string) string
	intstrstr func(int, string) string
	strintstr func(string, int) string
}

var op_add_fns = typefnmap{
	intintint: func(a, b int) int {
		return a + b
	},
	strstrstr: func(a, b string) string {
		return a + b
	},
	intstrstr: func(a int, b string) string {
		return strconv.Itoa(a) + b
	},
	strintstr: func(a string, b int) string {
		return a + strconv.Itoa(b)
	},
}

func (m *Machine) eval_op_dyn(scope *Scope, node ksyn.Node, fns typefnmap, depth int) (Value, error) {
	r := Value{}

	for i, ov := range node.Nodes {
		e_ov, err := m.eval(scope, ov, depth)
		if err != nil {
			return r, err
		}
		if i == 0 {
			r = e_ov
		} else {
			vt := e_ov.Value.Type
			rt := r.Value.Type

			if rt == parse.ValInt && vt == parse.ValInt && fns.intintint != nil {
				r.Value = parse.Value{
					Type: parse.ValInt,
					Int:  fns.intintint(r.Value.Int, e_ov.Value.Int),
				}
			} else if rt == parse.ValString && vt == parse.ValString && fns.strstrstr != nil {
				r.Value = parse.Value{
					Type: parse.ValString,
					Str:  fns.strstrstr(r.Value.Str, e_ov.Value.Str),
				}
			} else if rt == parse.ValInt && vt == parse.ValString && fns.intstrstr != nil {
				r.Value = parse.Value{
					Type: parse.ValString,
					Str:  fns.intstrstr(r.Value.Int, e_ov.Value.Str),
				}
			} else if rt == parse.ValString && vt == parse.ValInt && fns.strintstr != nil {
				r.Value = parse.Value{
					Type: parse.ValString,
					Str:  fns.strintstr(r.Value.Str, e_ov.Value.Int),
				}
			} else {
				return r, EvalError{
					Err: fmt.Errorf("Cannot %s types %s and %s", node.Op, rt, vt),
					Pos: ov.Pos,
				}
			}
		}
	}

	return r, nil
}*/
