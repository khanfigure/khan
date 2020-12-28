package khan

import (
	"errors"
)

type FuncType func(*Host) error

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

func (f *Function) SetID(id int) {
	f.id = id
}
func (f *Function) ID() int {
	return f.id
}
func (f *Function) Clone() Item {
	r := *f
	r.id = 0
	return &r
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

func (f *Function) Needs() []string {
	return nil
}
func (f *Function) Provides() []string {
	return nil
}

func (f *Function) Apply(host *Host) (itemStatus, error) {
	return itemUnchanged, f.Fn(host)
}
