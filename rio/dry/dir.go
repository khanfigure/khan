package dry

import (
	"path"
	"time"

	"khan.rip/rio/util"
)

func (host *Host) MkdirAll(fpath string) error {
	host.fsmu.Lock()
	defer host.fsmu.Unlock()

	file := host.fs[fpath]
	if file == nil && host.cascade != nil {
		f, err := host.cascade.Stat(fpath)
		if err != nil {
			if util.IsErrNotFound(err) {
				// Okay cool, we can make file here.
			} else {
				return err
			}
		} else {
			fi, err := util.ConvertStat(f)
			if err != nil {
				return err
			}

			file = &File{
				info: fi,
			}
			host.fs[fpath] = file
		}
	}
	if file != nil && file.info.Fisdir {
		// No-op, it already exists and is a directory
		return nil
	}

	if err := util.MkdirAll(host, fpath); err != nil {
		return err
	}

	// Simulate creating the file
	file = &File{
		info: &util.FileInfo{
			Fname:    path.Base(fpath),
			Fmode:    0700,
			Fmodtime: time.Now(),
			Fisdir:   true,
			Fuid:     host.uid,
			Fgid:     host.gid,
		},
	}
	host.fs[fpath] = file
	return nil
}
