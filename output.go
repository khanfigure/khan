package duck

import (
	"fmt"
	"time"
)

type outputter struct {
	start time.Time
}

func (o *outputter) StartItem(item Item) {
	o.start = time.Now()
}

func (o *outputter) FinishItem(item Item, err error) {
	fmt.Printf("%20s %s\n", format_duration(time.Since(o.start)), item.String())
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
	}

	return d.String()
}
