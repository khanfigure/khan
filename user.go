package khan

import (
	"fmt"
	"sort"

	"github.com/desops/khan/rio"
)

type User struct {
	Name string

	// Primary group. If not specified, user name is used
	Group string

	Uid uint32

	// Supplemental groups
	Groups []string

	Comment string // Gecos field

	Home  string
	Shell string

	// Password is the passord encrypted with libcrypt.
	// Password if blank will actually be set to "!". If "!", "!!", or "x" are found
	// in /etc/shadow, it will be translated to a blank password. If you want an actually
	// blank password (not safe) use BlankPassword: true (blank_password: true in yaml).
	Password      string
	BlankPassword bool `khan:"blank_password"`

	// TODO fancy /etc/shadow fields

	Delete bool

	id int
}

func (u *User) String() string {
	return fmt.Sprintf("%s/%d", u.Name, u.Uid)
}

func (u *User) SetID(id int) {
	u.id = id
}
func (u *User) ID() int {
	return u.id
}
func (u *User) Clone() Item {
	r := *u
	r.id = 0
	return &r
}
func (u *User) After() []string {
	if !u.Delete {
		grp := u.Group
		if grp == "" {
			grp = u.Name
		}
		afters := make([]string, len(u.Groups)+1)
		afters[0] = "group:" + grp
		for i, v := range u.Groups {
			afters[i+1] = "group:" + v
		}
		return afters
	}
	return nil
}
func (u *User) Before() []string {
	if u.Delete {
		grp := u.Group
		if grp == "" {
			grp = u.Name
		}
		afters := make([]string, len(u.Groups)+1)
		afters[0] = "-group:" + grp
		for i, v := range u.Groups {
			afters[i+1] = "-group:" + v
		}
		return afters
	}
	return nil
}
func (u *User) Provides() []string {
	if u.Delete {
		return []string{"-user:" + u.Name}
	} else {
		return []string{"user:" + u.Name}
	}
}

func (u *User) Apply(host *Host) (itemStatus, error) {
	usergroup := u.Group
	if usergroup == "" {
		usergroup = u.Name
	}

	userhome := u.Home
	if userhome == "" {
		userhome = "/home/" + u.Name
	}

	usershell := u.Shell
	if usershell == "" {
		info, err := host.rh.Info()
		if err != nil {
			return 0, err
		}
		switch info.OS {
		case "openbsd":
			usershell = "/bin/ksh"
		default:
			usershell = "/bin/bash"
		}
	}

	old, err := host.rh.User(u.Name)
	if err != nil {
		return 0, err
	}

	if u.Delete {
		if old == nil {
			return itemUnchanged, nil
		}
		if err := host.rh.DeleteUser(u.Name); err != nil {
			return 0, err
		}
		return itemDeleted, nil
	}

	v := &rio.User{
		Name:    u.Name,
		Uid:     u.Uid,
		Group:   usergroup,
		Groups:  u.Groups,
		Home:    userhome,
		Shell:   u.Shell,
		Comment: u.Comment,
	}

	if old == nil {
		if err := host.rh.CreateUser(v); err != nil {
			return 0, err
		}
		return itemCreated, nil
	}

	modified := false

	if old.Uid != u.Uid {
		modified = true
	}

	/*		oldpw := old.Password
			if oldpw == "" && !old.BlankPassword {
				oldpw = "!"
			}
			newpw := u.Password
			if newpw == "" && !u.BlankPassword {
				newpw = "!"
			}
			if oldpw != newpw {
				modified = true
				//fmt.Printf("~ user %s (password)\n", u.Name)
				if v.OS == "openbsd" {
					// wish openbsd had chpasswd :'(
					// this leaks the crypted password hash via process args.
					// TODO maybe just buckle down and learn the proper way to lock the master file
					// and modify it directly?
					if err := printExec(host, "usermod", "-p", newpw, u.Name); err != nil {
						return 0, err
					}
				} else {
					input := bytes.NewBuffer([]byte(u.Name + ":" + newpw + "\n"))
					if err := printExecStdin(host, input, "chpasswd", "-e"); err != nil {
						return 0, err
					}
				}
				old.Password = u.Password
				old.BlankPassword = u.BlankPassword
			}*/

	g1 := make([]string, len(old.Groups))
	g2 := make([]string, len(u.Groups))
	copy(g1, old.Groups)
	copy(g2, u.Groups)
	sort.Strings(g1)
	sort.Strings(g2)
	if len(g1) != len(g2) {
		modified = true
	} else if len(g1) > 0 {
		for i, gg := range g1 {
			if g2[i] != gg {
				modified = true
				break
			}
		}
	}

	if old.Group != usergroup ||
		old.Home != userhome ||
		old.Shell != usershell ||
		old.Comment != u.Comment {
		modified = true
	}

	if modified {
		if err := host.rh.UpdateUser(v); err != nil {
			return 0, err
		}
		return itemModified, nil
	}

	return itemUnchanged, nil
}
