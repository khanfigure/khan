package duck

import (
	"fmt"
	"os"
)

var (
	nextid int
	items  []Item
)

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
