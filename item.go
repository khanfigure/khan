package khan

import (
	"fmt"
	"runtime"
)

type metadata struct {
	source string
}

type itemStatus int

const (
	invalidItemStatus itemStatus = iota
	itemUnchanged
	itemCreated
	itemModified
	itemDeleted
)

func (s itemStatus) String() string {
	switch s {
	case itemUnchanged:
		return "unchanged"
	case itemCreated:
		return "created"
	case itemModified:
		return "modified"
	case itemDeleted:
		return "deleted"
	default:
		return fmt.Sprintf("invalidItemStatus(%d)", s)
	}
}

type Item interface {
	SetID(id int)
	ID() int

	Clone() Item
	String() string

	Apply(host *Host) (itemStatus, error)

	Provides() []string
	Needs() []string
}

type Validator interface {
	Validate() error
}
type StaticFiler interface {
	StaticFiles() []string
}

// Add to the default run context
func Add(add ...Item) {
	_, fn, line, _ := runtime.Caller(1)
	source := fmt.Sprintf("%s:%d", fn, line)
	if err := run.AddFromSource(source, add...); err != nil {
		panic(err)
	}
}

// Add to the default run context with explicit source path
func AddFromSource(source string, add ...Item) {
	if err := run.AddFromSource(source, add...); err != nil {
		panic(err)
	}
}
