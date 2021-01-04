package remote

import (
	"sync"

	"github.com/desops/khan/rio"

	"github.com/desops/sshpool"
)

type Host struct {
	pool    *sshpool.Pool
	connect string

	infomu sync.Mutex
	info   *rio.Info

	usersmu   sync.Mutex
	users     map[string]*rio.User
	groups    map[string]*rio.Group
	passwords map[string]*rio.Password
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
