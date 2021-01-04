package khan

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/desops/sshpool"

	"github.com/flosch/pongo2/v4"
)

var (
	errNeededItemFailed = errors.New("Needed item failed")
)

// Run is the context for an execution run, on one or more servers.
type Run struct {
	Dry     bool
	Diff    bool
	Verbose bool

	Pool  *sshpool.Pool
	Hosts []*Host

	assetfn func(string) (io.ReadCloser, error)

	sourceprefix string
	describe     string
	title        string

	out *outputter

	pongomu            sync.Mutex
	pongopackedset     *pongo2.TemplateSet
	pongopackedcontext pongo2.Context
	pongocachefiles    map[string]*pongo2.Template
	pongocachestrings  map[string]*pongo2.Template

	itemsmu   sync.Mutex
	initdone  bool
	inititems []*inititem // items added at init() time -- need more processing before they're valid
	items     []Item
	meta      map[int]*imeta
	nextid    int
	fences    map[string]*sync.Mutex
	befores   map[string][]string
	errors    map[string]error
}

type inititem struct {
	run    *Run
	item   Item
	source string
}

func (ii *inititem) WrapError(run *Run, err error) error {
	return fmt.Errorf("%s %s: %w", strings.TrimPrefix(ii.source, run.sourceprefix+"/"), ii.item, err)
}

type imeta struct {
	item   Item
	source string
	host   *Host
}

func (im *imeta) WrapError(run *Run, err error) error {
	return fmt.Errorf("%s %s on %s: %w", strings.TrimPrefix(im.source, run.sourceprefix+"/"), im.item, im.host.Name, err)
}

// Add will clone items for each configured host and add them to the run graph
func (r *Run) Add(add ...Item) error {
	_, fn, line, _ := runtime.Caller(1)
	source := fmt.Sprintf("%s:%d", fn, line)
	return r.AddFromSource(source, add...)
}

// AddFromSource is like Add but with explicit source code path
func (r *Run) AddFromSource(source string, add ...Item) error {
	r.itemsmu.Lock()
	defer r.itemsmu.Unlock()

	if !r.initdone {
		for _, item := range add {
			r.inititems = append(r.inititems, &inititem{
				item:   item,
				source: source,
			})
		}
		return nil
	}

	for _, item := range add {
		if item.ID() != 0 {
			return fmt.Errorf("Item already added: %v", item)
		}
		for _, host := range r.Hosts {
			c := item.Clone()
			if err := r.addHostItem(host, source, c); err != nil {
				return err
			}
		}
	}

	return nil
}

// always have itemsmu locked before calling this
func (r *Run) addHostItem(host *Host, source string, item Item) error {
	if item.ID() != 0 {
		return fmt.Errorf("Item cannot be added twice: %v", item)
	}

	r.nextid++
	id := r.nextid

	item.SetID(id)

	im := &imeta{
		source: source,
		host:   host,
		item:   item,
	}

	r.meta[id] = im
	r.items = append(r.items, item)

	// create fences for things item provides
	for _, p := range item.Provides() {
		p = host.Key() + "-" + p
		if _, ok := r.fences[p]; ok {
			return im.WrapError(r, fmt.Errorf("Duplicate provider of %#v", p))
		}
		r.fences[p] = &sync.Mutex{}
		r.fences[p].Lock()

		for _, bef := range item.Before() {
			bef = host.Key() + "-" + bef
			r.befores[bef] = append(r.befores[bef], p)
		}
	}

	return nil
}

func (r *Run) runinit() error {
	// Do some initialization for items queued up at init() time.
	// Now that we have a proper host list, we can clone the items
	// for each host.
	r.itemsmu.Lock()
	defer r.itemsmu.Unlock()

	r.initdone = true

	for _, iitem := range r.inititems {
		if iitem.item.ID() != 0 {
			return iitem.WrapError(r, fmt.Errorf("Item already added"))
		}
		for _, host := range r.Hosts {
			c := iitem.item.Clone()
			if err := r.addHostItem(host, iitem.source, c); err != nil {
				return err
			}
		}
	}

	r.inititems = nil

	return nil
}

func (r *Run) run() error {
	if err := r.runinit(); err != nil {
		return err
	}

	r.out = &outputter{}

	errs := make(chan error)

	type iexec struct {
		item Item
		im   *imeta
	}

	var (
		exec     []*iexec
		running  int
		executed = map[int]bool{}

		errors             int
		interesting_errors []error
		skipfailures       int
	)

	for {

		r.itemsmu.Lock()
		for _, item := range r.items {
			if executed[item.ID()] {
				continue
			}
			executed[item.ID()] = true
			exec = append(exec, &iexec{
				item: item,
				im:   r.meta[item.ID()],
			})
		}
		r.itemsmu.Unlock()

		for _, ex := range exec {
			running++
			go func(ex *iexec) {
				item := ex.item
				host := ex.im.host

				err := func() error {
					// be a little tricky here to allow fences to appear in the future
					for {
						var (
							mu      *sync.Mutex
							waiting string
						)
						r.itemsmu.Lock()
						waitlist := make([]string, 0, len(item.After()))
						for _, after := range item.After() {
							waitlist = append(waitlist, host.Key()+"-"+after)
						}
						for _, pr := range item.Provides() {
							for _, bef := range r.befores[host.Key()+"-"+pr] {
								waitlist = append(waitlist, bef)
							}
						}
						for _, n := range waitlist {
							m, ok := r.fences[n]
							if ok {
								mu = m
								waiting = n
								break
							}
						}
						r.itemsmu.Unlock()

						if mu == nil {
							break
						}

						//fmt.Println(item, "awaiting", waiting)

						mu.Lock()
						mu.Unlock()

						// if a task we need fails, we need to fail too.
						r.itemsmu.Lock()
						parenterr, ok := r.errors[waiting]
						r.itemsmu.Unlock()

						if !ok {
							// WTF, this should never happen
							return fmt.Errorf("Parent task error status not found")
						}
						if parenterr != nil {
							return errNeededItemFailed
						}
					}

					start := time.Now()
					status, err := item.Apply(host)
					if err != nil {
						err = ex.im.WrapError(r, err)
					}
					r.out.FinishItem(start, r, item, status, err)

					return err
				}()

				r.itemsmu.Lock()
				for _, p := range item.Provides() {
					p = host.Key() + "-" + p
					r.errors[p] = err
					mu, ok := r.fences[p]
					if ok {
						mu.Unlock()
						delete(r.fences, p)
					}
				}
				r.itemsmu.Unlock()

				errs <- err
			}(ex)
		}

		exec = nil

		if running == 0 {
			if r.Dry {
				fmt.Fprintln(os.Stderr, "No actions actually performed (dry run)")
			}
			if errors == 0 && skipfailures == 0 {
				return nil
			}
			fmt.Fprintf(os.Stderr, "%s─── %d failures ───%s\n", color(Red), errors, reset())
			for _, err := range interesting_errors {
				fmt.Fprintln(os.Stderr, err)
			}
			return fmt.Errorf("%d items failed (%d items skipped)", errors, skipfailures)
		}

		// wait for something to finish
		err := <-errs
		running--

		if err != nil {
			if err == errNeededItemFailed {
				skipfailures++
			} else {
				interesting_errors = append(interesting_errors, err)
				errors++
			}
		} else {
			// success!
		}
	}
}
