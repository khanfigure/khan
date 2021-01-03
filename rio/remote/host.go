package remote

import (
	"sync"

	hhost "github.com/desops/khan/host"

	"github.com/desops/sshpool"
)

type Host struct {
	pool    *sshpool.Pool
	connect string

	infomu sync.Mutex
	info   *hhost.Info

	usersmu   sync.Mutex
	users     map[string]*hhost.User
	groups    map[string]*hhost.Group
	passwords map[string]*hhost.Password
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
