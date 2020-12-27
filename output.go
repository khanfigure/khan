package khan

import (
	"fmt"
	"strings"
	"time"

	"github.com/yobert/progress"
)

type outputter struct {
	start time.Time
	bar   *progress.Bar
}

func (o *outputter) StartItem(r *Run, item Item) {
	o.start = time.Now()
}

func (o *outputter) FinishItem(r *Run, item Item, status itemStatus, err error) {
	if err == nil && status == itemUnchanged && !r.verbose {
		return
	}

	d := time.Since(o.start)
	dc := ""
	ds := format_duration(d)
	if d > time.Millisecond*100 {
		dc = color(Red)
	}

	typ := strings.ToLower(strings.TrimPrefix(fmt.Sprintf("%T", item), "*khan."))

	s := status.String()
	if err != nil {
		s = "error"
	}

	msg := fmt.Sprintf("%s%8s%s │ %-10s │ %-10s │ %s", dc, ds, reset(), typ, s, item.String())

	if o.bar != nil {
		o.bar.Println(msg)
	} else {
		fmt.Println(msg)
	}
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
