package remote

import (
	"github.com/desops/sshpool"
)

type Host struct {
	pool    *sshpool.Pool
	connect string
}

func (host *Host) String() string {
	return "ssh " + host.connect
}

func New(pool *sshpool.Pool, connect string) *Host {
	return &Host{
		pool:    pool,
		connect: connect,
	}
}
