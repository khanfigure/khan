package local

import (
	"sync"

	"khan.rip/rio"
)

type Host struct {
	verbose bool

	// cache
	usersmu   sync.Mutex
	users     map[string]*rio.User
	groups    map[string]*rio.Group
	passwords map[string]*rio.Password

	tmpdirmu sync.Mutex
	tmpdir   string
}

func (host *Host) String() string {
	return "local"
}

func (host *Host) SetVerbose() {
	host.verbose = true
}

func New() *Host {
	return &Host{}
}
