package dry

import (
	"fmt"
	"sync"

	"github.com/desops/khan/host"
)

type Host struct {
	// the user/group to emulate operations as
	uid uint32
	gid uint32

	cascade host.Host

	fsmu sync.Mutex
	fs   map[string]*File
}

func (host *Host) String() string {
	return fmt.Sprintf("virtual host (uid %d gid %d) cascade → %s", host.uid, host.gid, host.cascade)
}

func New(uid, gid uint32, cascade host.Host) *Host {
	return &Host{
		uid:     uid,
		gid:     gid,
		cascade: cascade,
		fs:      map[string]*File{},
	}
}
