package rio

import (
	"github.com/desops/sshpool"
)

type Config struct {
	// Pool is an SSH connection pool. If nil, local mode is used on the current host.
	Pool *sshpool.Pool
}
