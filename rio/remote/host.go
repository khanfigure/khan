package remote

import (
	"sync"

	"khan.rip/rio"

	"github.com/desops/sshpool"
)

type Host struct {
	verbose bool

	pool    *sshpool.Pool
	connect string

	infomu sync.Mutex
	info   *rio.Info

	usersmu   sync.Mutex
	users     map[string]*rio.User
	groups    map[string]*rio.Group
	passwords map[string]*rio.Password

	tmpdirmu sync.Mutex
	tmpdir   string
}

func (host *Host) String() string {
	return "ssh " + host.connect
}

func (host *Host) SetVerbose() {
	host.verbose = true
}

func New(pool *sshpool.Pool, connect string) *Host {
	return &Host{
		pool:    pool,
		connect: connect,
	}
}
