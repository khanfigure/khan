package remote

import (
	"sync"

	"github.com/desops/khan/host"

	"github.com/desops/sshpool"
)

type Host struct {
	pool    *sshpool.Pool
	connect string

	infomu sync.Mutex
	info   *host.Info
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
