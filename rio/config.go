package rio

import (
	"github.com/desops/sshpool"
)

type Config struct {
	// Pool is an SSH connection pool. If nil, local mode is used on the current host.
	Pool *sshpool.Pool

	// Host to connect to. Leave blank for local mode.
	Host string

	// User to switch to. If blank, currently running user is used, or default user of SSH
	// connection for remote connections. If specified, sudo will be used to attempt to
	// switch to that user.
	Sudo string

	Verbose bool
}
