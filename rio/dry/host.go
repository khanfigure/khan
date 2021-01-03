package dry

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/desops/khan/rio"
)

type Host struct {
	// the user/group to emulate operations as
	uid uint32
	gid uint32

	cascade rio.Host

	fsmu sync.Mutex
	fs   map[string]*File

	usersmu   sync.Mutex
	users     map[string]*rio.User
	groups    map[string]*rio.Group
	passwords map[string]*rio.Password
}

func (host *Host) String() string {
	return fmt.Sprintf("virtual host (uid %d gid %d) cascade → %s", host.uid, host.gid, host.cascade)
}

func New(uid, gid uint32, cascade rio.Host) *Host {
	return &Host{
		uid:       uid,
		gid:       gid,
		cascade:   cascade,
		fs:        map[string]*File{},
		users:     map[string]*rio.User{},
		groups:    map[string]*rio.Group{},
		passwords: map[string]*rio.Password{},
	}
}

func (host *Host) Info() (*rio.Info, error) {
	if host.cascade != nil {
		return host.cascade.Info()
	}

	// TODO let you customize this
	return &rio.Info{
		Hostname: "dry",
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
	}, nil
}
