package duck

import (
	"fmt"
	"os"
)

var (
	nextid int
	items  []Item
)

type File struct {
	Path    string
	Content string
	User    User

	id int
}

func (f *File) setID(id int) {
	f.id = id
}
func (f *File) getID() int {
	return f.id
}
func (f *File) apply() error {
	fmt.Println("setting", f.Path, "to", f.Content)
	return nil
}

type User struct {
	Name     string
	Password string

	Uid        int
	Gid        int
	GroupNames []string
}
type Group struct {
	Name string
	Gid  int
}

type Package struct {
	Name string
}

type Item interface {
	setID(id int)
	getID() int
	apply() error
}

func Add(add ...Item) {
	// adding something gives it a unique id

	for _, item := range add {
		if item.getID() != 0 {
			// already added
			continue
		}
		nextid++
		item.setID(nextid)
		items = append(items, item)
	}
}

func Apply() error {
	var firsterr error
	for _, item := range items {
		if err := item.apply(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			if firsterr != nil {
				firsterr = err
			}
		}
	}
	return firsterr
}

/*func Main(items ...Item) {
	exit := 0
	for _, item := range items {
		if err := apply(item); err != nil {
			fmt.Println(err)
			exit = 1
		}
	}
	os.Exit(exit)
}

func apply(item Item) error {
	switch v := item.(type) {
	case []Item:
		for _, i := range v {
			if err := apply(i); err != nil {
				return err
			}
		}
	case File:
		fmt.Println("setting", v.Path, "to", v.Content)
	case User:
		fmt.Println("creating unix user", v.Name)
	case Group:
		fmt.Println("creating unix group", v.Name)
	case Package:
		fmt.Println("installing package", v.Name)
	default:
		return fmt.Errorf("Unhandled item type %T", item)
	}
	return nil
}*/
