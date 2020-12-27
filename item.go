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
	setID(id int)
	getID() int
	apply(*Run) (itemStatus, error)
	String() string

	provides() []string
	needs() []string
}

type Validator interface {
	Validate() error
}
type StaticFiler interface {
	StaticFiles() []string
}

func Add(add ...Item) {
	_, fn, line, _ := runtime.Caller(1)
	source := fmt.Sprintf("%s:%d", fn, line)
	if err := run.AddFromSource(source, add...); err != nil {
		panic(err)
	}
}

func AddFromSource(source string, add ...Item) {
	if err := run.AddFromSource(source, add...); err != nil {
		panic(err)
	}
}
