package keval

import (
	"fmt"
)

type Value struct {
	V     interface{}
	Scope *Scope
}

func (v Value) String() string {
	return fmt.Sprintf("%#v", v.V)
}
