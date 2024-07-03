package dry

import (
	"log"

	"khan.rip/rio"
)

func (host *Host) Exec(cmd *rio.Cmd) error {
	if cmd.ReadOnly && host.cascade != nil {
		// actually execute it.  anything that doesn't have side effects, we want to run,
		// so we can get potential errors during a dry run wherever possible.
		return host.cascade.Exec(cmd)
	}

	if host.verbose {
		log.Println(host, cmd)
	}
	return nil
}
