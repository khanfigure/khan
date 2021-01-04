package dry

import (
	"os"
	"syscall"
)

func (host *Host) Stat(fpath string) (os.FileInfo, error) {
	host.fsmu.Lock()
	file := host.fs[fpath]
	if file == nil && host.cascade != nil {
		// we don't want to hold this lock while SSH does its thing
		host.fsmu.Unlock()
		return host.cascade.Stat(fpath)
	}
	defer host.fsmu.Unlock()

	if file == nil || file.info == nil {
		return nil, &os.PathError{Op: "stat", Path: fpath, Err: syscall.ENOENT}
	}

	return file.info, nil
}
