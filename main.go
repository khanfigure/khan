package duck

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	describe     string
	sourceprefix string
)

func SetSourcePrefix(s string) {
	sourceprefix = s
}
func SetDescribe(s string) {
	describe = s
}

func Apply() error {
	r := &run{
		assetfn: mainassetfn,
	}

	flag.BoolVar(&r.dry, "d", false, "Dry run; Don't make any changes")
	flag.BoolVar(&r.diff, "D", false, "Show full diff of file content changes")
	flag.BoolVar(&r.verbose, "v", false, "Be more verbose")

	flag.StringVar(&r.ssh, "ssh", "", "SSH mode connection string (host, user@host, or user@host:port)")

	flag.Parse()

	title := "░░░ Configuration " + brightcolor(Yellow) + describe + reset() + " "

	if r.dry {
		title += color(Green) + "dry running"
	} else {
		title += color(Red) + "executing"
	}
	title += reset()
	if r.ssh == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		title += " " + hostname
	} else {
		title += " " + r.ssh
	}
	title += " ░░░"
	fmt.Println(title)

	out := &outputter{}
	r.out = out

	for _, item := range items {
		out.StartItem(item)

		err := item.apply(r)

		out.FinishItem(item, err)

		if err == ErrUnchanged {
			err = nil
		}

		if err != nil {
			md := meta[item.getID()]

			// wrap the error with its source
			err = fmt.Errorf("%s %w", strings.TrimPrefix(md.source, sourceprefix+"/"), err)

			return err
		}
	}

	return nil
}
