package khan

import (
	"fmt"
	"strings"
	"time"
)

type outputter struct {
}

func (o *outputter) Active(r *Run, item Item, status Status) {
	color := status.Color()

	boldcolor := color
	boldcolor.Bold = true

	typ := strings.ToLower(strings.TrimPrefix(fmt.Sprintf("%T", item), "*khan."))

	fmt.Println(boldcolor.Wrap(fmt.Sprintf("%8s %s", status.ActiveString(), typ)) + " " + color.Wrap(item.String()))
}
func (o *outputter) FinishItem(start time.Time, r *Run, item Item, status Status, err error) {
	if err == nil && status == Unchanged && !r.Verbose {
		return
	}

	d := time.Since(start)

	typ := strings.ToLower(strings.TrimPrefix(fmt.Sprintf("%T", item), "*khan."))

	color := status.Color()
	if err != nil {
		color.Color = Red
	}

	boldcolor := color
	boldcolor.Bold = true

	msg := boldcolor.Wrap(fmt.Sprintf("%8s %s", status.String(), typ)) + " " + color.Wrap(item.String())

	msg += " in " + color_duration(d).Wrap(format_duration(d))

	if err != nil {
		msg += " failed: " + err.Error()
	}

	//	if o.bar != nil {
	//		o.bar.Println(msg)
	//	} else {
	//	}

	fmt.Println(msg)
	_ = msg
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
	} else if d > time.Microsecond {
		d = d / time.Microsecond * time.Microsecond
	}

	return d.String()
}

func color_duration(d time.Duration) Color {
	var dc Color
	if d < time.Millisecond {
		dc.Dim = true
	} else if d > time.Millisecond*100 {
		dc.Dim = true
		dc.Color = Red
	} else if d > time.Second {
		dc.Dim = false
		dc.Color = Red
	}
	return dc
}
