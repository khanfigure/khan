package khan

import (
	"fmt"
	"runtime"
)

type Status int

const (
	InvalidStatus Status = iota
	Unchanged
	Created
	Modified
	Deleted
)

func (s Status) String() string {
	switch s {
	case Unchanged:
		return "unchanged"
	case Created:
		return "created"
	case Modified:
		return "modified"
	case Deleted:
		return "deleted"
	default:
		return fmt.Sprintf("invalidItemStatus(%d)", s)
	}
}
func (s Status) ActiveString() string {
	switch s {
	case Created:
		return "creating"
	case Modified:
		return "updating"
	case Deleted:
		return "deleting"
	default:
		return fmt.Sprintf("invalidItemStatus(%d)", s)
	}
}

func (s Status) Color() Color {
	switch s {
	case Created:
		return Color{Color: Green}
	case Modified:
		return Color{Color: Cyan}
	case Deleted:
		return Color{Color: Yellow}
	}
	return Color{}
}

type Item interface {
	SetID(id int)
	ID() int

	Clone() Item
	String() string

	Apply(host *Host) (Status, error)

	Provides() []string
	After() []string
	Before() []string
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
	if err := defaultrun.AddFromSource(source, add...); err != nil {
		panic(err)
	}
}

// Add to the default run context with explicit source path
func AddFromSource(source string, add ...Item) {
	if err := defaultrun.AddFromSource(source, add...); err != nil {
		panic(err)
	}
}
