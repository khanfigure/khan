package dry

import (
	"fmt"

	hhost "github.com/desops/khan/host"
)

func (host *Host) Exec(cmd *hhost.Cmd) error {
	fmt.Println(host, cmd)

	if cmd.ReadOnly && host.cascade != nil {
		// actually execute it.  anything that doesn't have side effects, we want to run,
		// so we can get potential errors during a dry run wherever possible.
		return host.cascade.Exec(cmd)
	}

	fmt.Println(host, cmd, "[NOT EXECUTED]")
	return nil
}
