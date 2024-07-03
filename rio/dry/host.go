package dry

import (
	"fmt"
	"runtime"
	"sync"

	"khan.rip/rio"
)

type Host struct {
	verbose bool

	// the user/group to emulate operations as
	uid uint32
	gid uint32

	cascade rio.Host

	fsmu    sync.Mutex
	fs      map[string]*File
	tmpdir  string
	tmpfile int

	usersmu   sync.Mutex
	users     map[string]*rio.User
	groups    map[string]*rio.Group
	passwords map[string]*rio.Password
}

func (host *Host) String() string {
	if host.cascade != nil {
		return fmt.Sprintf("dry run (%d:%d) %s", host.uid, host.gid, host.cascade)
	}
	return fmt.Sprintf("dry run (%d:%d)", host.uid, host.gid)
}
func (host *Host) SetVerbose() {
	host.verbose = true
}

func (host *Host) debug() {
	for fpath, file := range host.fs {
		fmt.Println(fpath)
		fmt.Println("\t" + file.String())
	}
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
