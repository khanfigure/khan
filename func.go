package khan

import (
	"errors"
)

type FuncType func(*Run) error

type Function struct {
	Fn FuncType
	id int
}

func Func(fn FuncType) Item {
	return &Function{
		Fn: fn,
	}
}

func (f *Function) String() string {
	return "function"
}

func (f *Function) setID(id int) {
	f.id = id
}
func (f *Function) getID() int {
	return f.id
}

func (f *Function) Validate() error {
	if f.Fn == nil {
		return errors.New("Function reference is required")
	}
	return nil
}

func (f *Function) StaticFiles() []string {
	return nil
}

func (f *Function) needs() []string {
	return nil
}
func (f *Function) provides() []string {
	return nil
}

func (f *Function) apply(r *Run) (itemStatus, error) {
	return itemUnchanged, f.Fn(r)
}
