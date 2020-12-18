package duck

import (
	"fmt"
	"strings"
	"time"
)

type outputter struct {
	start time.Time
}

func (o *outputter) StartItem(r *run, item Item) {
	o.start = time.Now()
}

func (o *outputter) FinishItem(r *run, item Item, status itemStatus, err error) {
	if status == itemUnchanged && !r.verbose {
		return
	}

	d := time.Since(o.start)
	dc := ""
	ds := format_duration(d)
	if d > time.Millisecond*100 {
		dc = color(Red)
	}

	typ := strings.ToLower(strings.TrimPrefix(fmt.Sprintf("%T", item), "*duck."))

	fmt.Printf("%s%8s%s │ %-10s │ %-10s │ %s\n", dc, ds, reset(), typ, status, item.String())
}

func (o *outputter) Flush() {
}

func format_duration(d time.Duration) string {
	if d > time.Hour {
		d = d / time.Minute * time.Minute
	} else if d > time.Minute {
		d = d / time.Second * time.Second
	} else if d > time.Second {
		d = d / (time.Millisecond * 100) * time.Millisecond * 100
	} else if d > time.Millisecond {
		d = d / time.Millisecond * time.Millisecond
	} else {
		d = d / time.Microsecond * time.Microsecond
	}

	return d.String()
}
