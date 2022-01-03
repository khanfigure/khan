package keval

import (
	"khan.rip/ksyn"
)

func (m *Machine) eval_for_in(scope *Scope, node ksyn.ForInStmt, depth int) (Value, error) {
	list, err := m.eval(scope, node.Expr, depth)
	if err != nil {
		return Value{}, err
	}

	each := func(v Value) (Value, error) {
		escope := &Scope{
			parent: scope,
			symbols: map[string]Value{
				node.Ident.Ident: v,
			},
		}

		return m.eval(escope, node.Block, depth)
	}

	switch lv := list.V.(type) {
	case []interface{}:
		for _, v := range lv {
			_, ee := each(Value{V: v})
			if ee != nil {
				return Value{}, ee
			}
		}
		return Value{}, nil
	case map[string]interface{}:
		for k := range lv {
			_, ee := each(Value{V: k})
			if ee != nil {
				return Value{}, ee
			}
		}
		return Value{}, nil
	}

	return Value{}, ksyn.Errorf(node.Expr, "Cannot iterate a %T", list.V)
}
