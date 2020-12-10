package duck

import (
	"fmt"
	"io"
	"os"
	"sort"
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
	apply(*run) error
}

type Validator interface {
	Validate() error
}
type StaticFiler interface {
	StaticFiles() []string
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

func Apply(assetfn func(string) (io.Reader, error)) error {
	r := &run{
		assetfn: assetfn,
		stats:   map[string]int{},
	}

	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" {
			fmt.Fprintf(os.Stderr, `Usage: %s [options]

    -d  --dry      Dry run, don't make any changes
    -D  --diff     Show full diff of file content changes
    -v  --verbose  Be more verbose

    -h --help      This text

`, os.Args[0])
			os.Exit(1)
		}

		if arg == "-d" || arg == "--dry" {
			r.dry = true
		} else if arg == "-D" || arg == "--diff" {
			r.diff = true
		} else if arg == "-v" || arg == "--verbose" {
			r.verbose = true
		} else {
			fmt.Fprintf(os.Stderr, "Invalid parameter %#v. Try -h / --help for usage.", arg)
			os.Exit(1)
		}
	}

	if r.dry {
		fmt.Println("Dry run mode: Nothing will be changed")
	} else {
		fmt.Println("Executing changes")
	}
	fmt.Println()

	var firsterr error
	for _, item := range items {
		if err := item.apply(r); err != nil {
			fmt.Fprintln(os.Stderr, item, err)
			if firsterr == nil {
				firsterr = err
			}
		}
	}
	fmt.Println("────────────")
	var keys []string
	for k := range r.stats {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := r.stats[k]
		fmt.Printf("%-20s %#v\n", k, v)
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
