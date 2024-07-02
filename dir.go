package khan

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"

	"khan.rip/rio/util"
)

type Dir struct {
	Path string `khan:"path,shortkey"`

	User  string
	Group string
	Mode  os.FileMode

	Delete bool

	id int
}

func (d *Dir) String() string {
	return d.Path
}

func (d *Dir) SetID(id int) {
	d.id = id
}
func (d *Dir) ID() int {
	return d.id
}
func (d *Dir) Clone() Item {
	r := *d
	r.id = 0
	return &r
}

func (d *Dir) Validate() error {
	if d.Path == "" {
		return errors.New("Dir path is required")
	}
	return nil
}

func (d *Dir) StaticFiles() []string {
	return nil
}

func (d *Dir) After() []string {
	if d.Delete {
		return nil
	}
	var afters []string
	if d.User != "" {
		afters = append(afters, "user:"+d.User)
	}
	if d.Group != "" {
		afters = append(afters, "group:"+d.Group)
	}
	return afters
}
func (d *Dir) Before() []string {
	return nil
}
func (d *Dir) Provides() []string {
	return []string{"path:" + d.Path}
}

func (d *Dir) Apply(host *Host) (itemStatus, error) {
	if d.Delete {
		_, err := host.rh.Stat(d.Path)
		if err != nil && util.IsErrNotFound(err) {
			return itemUnchanged, nil
		}
		if err != nil {
			return 0, err
		}
		if err := host.rh.Remove(d.Path); err != nil {
			return 0, err
		}
		return itemDeleted, nil
	}

	mode := d.Mode
	if mode == 0 {
		mode = 0755
	}

	ustr := d.User
	if ustr == "" {
		at := strings.IndexByte(host.Host, '@')
		if host.SSH && at > -1 {
			ustr = host.Host[:at]
		} else {
			osu, err := user.Current()
			if err != nil {
				return 0, err
			}
			ustr = osu.Username
		}
	}
	if ustr == "" {
		return 0, fmt.Errorf("Cannot determine user for managed directory %v", d)
	}

	user, err := host.rh.User(ustr)
	if err != nil {
		return 0, err
	}

	gstr := d.Group
	if gstr == "" {
		gstr = user.Group
	}

	if gstr == "" {
		return 0, fmt.Errorf("Cannot determine group for managed directory %v", d)
	}

	group, err := host.rh.Group(gstr)
	if err != nil {
		return 0, err
	}

	var (
		uid     uint32
		gid     uint32
		wantuid uint32
		wantgid uint32
	)

	if user == nil {
		return 0, fmt.Errorf("Unknown user %#v", ustr)
	}
	if group == nil {
		return 0, fmt.Errorf("Unknown group %#v", gstr)
	}

	wantuid = user.Uid
	wantgid = group.Gid

	created := false
	fpath := d.Path

	fi, err := host.rh.Stat(fpath)
	if err != nil {

		if util.IsErrNotFound(err) {
			// Create the directory
			created = true

			mkerr := host.rh.MkdirAll(fpath)
			if mkerr != nil {
				return 0, mkerr
			}

			// Now re-stat. It had better succeed.
			fi, err = host.rh.Stat(fpath)
			if err != nil {
				return 0, err
			}

			// Continue below to apply permissions.
		} else {
			return 0, err
		}
	}

	ufi, err := util.ConvertStat(fi)
	if err != nil {
		return 0, err
	}

	uid = ufi.Fuid
	gid = ufi.Fgid

	status := itemUnchanged

	if wantuid != uid || wantgid != gid {
		//fmt.Printf("wantuid %d wantgid %d uid %d gid %d\n", wantuid, wantgid, uid, gid)
		status = itemModified

		if err := host.rh.Chown(fpath, wantuid, wantgid); err != nil {
			return 0, err
		}
	}

	if fi.Mode()&util.S_justmode != mode {
		//fmt.Printf("current: %o , masked %o , want: %o\n", uint32(fi.Mode()), uint32(fi.Mode())&util.S_justmode, mode)
		status = itemModified

		if err := host.rh.Chmod(fpath, mode); err != nil {
			return 0, err
		}
	}

	if created {
		status = itemCreated
	}

	return status, nil
}
