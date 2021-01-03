package local

import (
	"sync"

	hhost "github.com/desops/khan/host"
)

type Host struct {
	// cache
	usersmu   sync.Mutex
	users     map[string]*hhost.User
	groups    map[string]*hhost.Group
	passwords map[string]*hhost.Password
}

func (host *Host) String() string {
	return "local"
}

func New() *Host {
	return &Host{}
}
