package duck

import (
	"fmt"
	"runtime"
)

type metadata struct {
	source string
}

var (
	nextid int
	items  []Item
	meta   map[int]*metadata
)

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
	apply(*run) (itemStatus, error)
	String() string
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
	AddFromSource(source, add...)
}

func AddFromSource(source string, add ...Item) {
	if meta == nil {
		meta = make(map[int]*metadata)
	}

	// adding something gives it a unique id

	for _, item := range add {
		if item.getID() != 0 {
			// already added
			continue
		}
		nextid++
		item.setID(nextid)
		items = append(items, item)
		meta[nextid] = &metadata{
			source: source,
		}
	}
}
