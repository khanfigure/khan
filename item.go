package duck

import (
	"errors"
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

	ErrUnchanged = errors.New("unchanged")
)

type Item interface {
	setID(id int)
	getID() int
	apply(*run) error
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
