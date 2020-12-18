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

	// not implemented yet
	//flag.StringVar(&r.ssh, "ssh", "", "SSH mode connection string (host, user@host, or user@host:port)")

	flag.Parse()

	title := "░░░ Configuration " + brightcolor(Yellow) + describe + reset() + " "

	if r.dry {
		title += color(Green) + "dry run"
	} else {
		title += color(Red) + "execute"
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

	total := len(items)
	finished := 0

	defer func() {
		fmt.Printf("%d/%d things up to date\n", finished, total)
	}()

	for _, item := range items {
		out.StartItem(r, item)

		status, err := item.apply(r)

		out.FinishItem(r, item, status, err)

		if err != nil {
			// wrap the error with its source
			md := meta[item.getID()]
			err = fmt.Errorf("%s %w", strings.TrimPrefix(md.source, sourceprefix+"/"), err)
			return err
		}

		finished++
	}

	return nil
}
