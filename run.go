package khan

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"

	"github.com/desops/khan/rio"

	"github.com/flosch/pongo2/v4"
)

// The context for an execution on a server
type Run struct {
	dry     bool
	diff    bool
	verbose bool

	host string
	user string

	rioconfig *rio.Config

	userCacheMu sync.Mutex
	userCache   map[string]*User
	uidCache    map[int]*User
	groupCache  map[string]*Group
	gidCache    map[int]*Group
	bsdmode     bool

	assetfn func(string) (io.ReadCloser, error)

	out *outputter

	pongopackedset     *pongo2.TemplateSet
	pongopackedcontext pongo2.Context
	pongocachefiles    map[string]*pongo2.Template
	pongocachestrings  map[string]*pongo2.Template

	itemsmu   sync.Mutex
	items     []Item
	meta      map[int]*metadata
	nextid    int
	moreitems chan ([]Item)
}

/*func (r *Run) addStat(stat string) {
	r.statsMu.Lock()
	r.stats[stat] = r.stats[stat] + 1
	r.statsMu.Unlock()
}*/

func (r *Run) Add(add ...Item) error {
	_, fn, line, _ := runtime.Caller(1)
	source := fmt.Sprintf("%s:%d", fn, line)
	return r.AddFromSource(source, add...)
}

func (r *Run) AddFromSource(source string, add ...Item) error {
	// careful with this defer ordering:
	// we want to send on moreitems after unlocking itemsmu,
	// but we want to make sure we have the lock first to
	// ensure the moreitems channel has been created.
	r.itemsmu.Lock()
	if r.moreitems != nil {
		defer func() {
			r.moreitems <- add
		}()
	}
	defer r.itemsmu.Unlock()

	if r.meta == nil {
		r.meta = make(map[int]*metadata)
	}

	// adding something gives it a unique id

	for _, item := range add {
		if item.getID() != 0 {
			// already added?
			// maybe we'll allow this later, but avoid for now
			return fmt.Errorf("Item cannot be added twice")
			//continue
		}
		r.nextid++
		item.setID(r.nextid)
		r.items = append(r.items, item)
		r.meta[r.nextid] = &metadata{
			source: source,
		}
	}

	return nil
}

func (r *Run) run() error {

	r.out = &outputter{}

	var wg sync.WaitGroup

	var metamu sync.Mutex
	fences := map[string]*sync.Mutex{}
	errors := map[string]error{}
	needed := map[string]bool{}

	wraperr := func(item Item, err error) error {
		md := r.meta[item.getID()]
		return fmt.Errorf("%s %w", strings.TrimPrefix(md.source, sourceprefix+"/"), err)
	}

	additem := func(item Item) error {
		for _, n := range item.needs() {
			needed[n] = true
		}
		for _, p := range item.provides() {
			if _, ok := fences[p]; ok {
				return wraperr(item, fmt.Errorf("Duplicated provider of %#v", p))
			}
			fences[p] = &sync.Mutex{}
			fences[p].Lock()
		}
		return nil
	}

	r.itemsmu.Lock()
	r.moreitems = make(chan []Item)
	for _, item := range r.items {
		if err := additem(item); err != nil {
			return err
		}
		wg.Add(1)
	}
	origitems := r.items
	r.itemsmu.Unlock()

	iexec := make(chan Item)

	// read for incoming new items added at runtime.
	// since the channel is blocking, the Add won't return until
	// the new item is handled here, thus ensuring it gets added
	// to the waitgroup.
	go func() {
		for {
			items, ok := <-r.moreitems
			if !ok {
				return
			}
			//fmt.Println("additional item locking")
			metamu.Lock()
			//fmt.Println("additional item locking done")
			for _, item := range items {
				err := additem(item)
				if err != nil {
					panic(err) // yuck: figure this out
				}
				wg.Add(1)
			}
			metamu.Unlock()

			for _, item := range items {
				iexec <- item
			}
		}
	}()

	errs := make(chan error)

	go func() {
		//fmt.Println("wg.Wait()")
		wg.Wait()
		//fmt.Println("closing errs")
		close(errs)
	}()

	go func() {
		for _, item := range origitems {
			iexec <- item
		}
	}()

	go func() {
		for {
			item, ok := <-iexec
			if !ok {
				return
			}

			//fmt.Println(item)
			go func(item Item) {
				err := func() error {

					// be a little tricky here to allow fences to appear in the future
					for {
						var (
							mu      *sync.Mutex
							waiting string
						)
						metamu.Lock()
						for _, n := range item.needs() {
							m, ok := fences[n]
							if ok {
								mu = m
								waiting = n
								break
							}
						}
						metamu.Unlock()

						if mu == nil {
							break
						}

						//fmt.Println(item, "awaiting", waiting)
						_ = waiting
						mu.Lock()
						mu.Unlock()
					}

					r.out.StartItem(r, item)
					status, err := item.apply(r)
					r.out.FinishItem(r, item, status, err)

					if err != nil {
						// wrap the error with its source
						md := r.meta[item.getID()]
						err = fmt.Errorf("%s %w", strings.TrimPrefix(md.source, sourceprefix+"/"), err)
						return err
					}

					//		if !r.dry || status == itemUnchanged {
					//			finished++
					//		}

					return nil
				}()

				metamu.Lock()
				for _, p := range item.provides() {
					errors[p] = err
					mu, ok := fences[p]
					if ok {
						mu.Unlock()
						delete(fences, p)
					}
				}
				metamu.Unlock()

				//r.out.bar.Next()
				errs <- err
				wg.Done()
			}(item)
		}

	}()

	for {
		err, ok := <-errs
		if !ok {
			// done!
			return nil
		}
		//fmt.Println("err", err)
		if err != nil {
			return err
		}
	}
}
