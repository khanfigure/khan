package dry

import (
	"fmt"
	"os"
	"path"
	"time"

	"khan.rip/rio/util"
)

func (host *Host) TmpFile() (string, error) {
	tmpdir, err := host.TmpDir()
	if err != nil {
		return "", err
	}

	host.fsmu.Lock()
	defer host.fsmu.Unlock()

	fpath := fmt.Sprintf("%s/%d", tmpdir, host.tmpfile)
	for host.fs[fpath] != nil {
		host.tmpfile++
		fpath = fmt.Sprintf("%s/%d", tmpdir, host.tmpfile)
	}
	host.tmpfile++

	return fpath, nil
}

func (host *Host) TmpDir() (string, error) {
	host.fsmu.Lock()
	defer host.fsmu.Unlock()

	if host.tmpdir != "" {
		return host.tmpdir, nil
	}

	i := os.Getpid()
	fpath := fmt.Sprintf("/tmp/tmpkhan_%d", i)
	for host.fs[fpath] != nil {
		i++
		fpath = fmt.Sprintf("/tmp/tmpkhan_%d", i)
	}

	if err := util.Mkdir(host, fpath); err != nil {
		return "", err
	}
	file := &File{
		info: &util.FileInfo{
			Fname:    path.Base(fpath),
			Fsize:    0,
			Fmode:    0700,
			Fmodtime: time.Now(),
			Fisdir:   true,
			Fuid:     host.uid,
			Fgid:     host.gid,
		},
	}
	host.fs[fpath] = file
	host.tmpdir = fpath
	return fpath, nil
}

func (host *Host) Cleanup() error {
	host.fsmu.Lock()
	defer host.fsmu.Unlock()

	if host.tmpdir == "" {
		return nil
	}
	if err := util.RemoveAll(host, host.tmpdir); err != nil {
		return err
	}
	host.fs[host.tmpdir] = &File{}
	return nil
}
