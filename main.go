package duck

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
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
	buf, err := ioutil.ReadFile(f.Path)
	//fmt.Printf("path %#v err %#v compare %#v buf %#v content %#v\n", f.Path, err, bytes.Compare(buf, []byte(f.Content)), string(buf), f.Content)
	if err == nil && bytes.Compare(buf, []byte(f.Content)) == 0 {
		// no change
		return nil
	}
	if err != nil && iserrnotfound(err) {
		fmt.Printf("+ %s\n", f.Path)
	} else {
		reason := "content updated"
		if err != nil {
			reason = err.Error()
		}
		fmt.Printf("~ %s (%s)\n", f.Path, reason)
	}
	fh, err := os.Create(f.Path)
	if err != nil {
		return err
	}
	defer fh.Close()
	if _, err := fh.Write([]byte(f.Content)); err != nil {
		return err
	}
	if err := fh.Close(); err != nil {
		return err
	}
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

func iserrnotfound(err error) bool {
	// TODO do this better
	v, ok := err.(*os.PathError)
	if ok && v != nil && v.Err == syscall.ENOENT {
		return true
	}
	return false
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
