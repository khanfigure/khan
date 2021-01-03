package khan

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type User struct {
	Name string

	// Primary group. If not specified, user name is used
	Group string

	Uid uint32

	// Supplemental groups
	Groups []string

	Gecos string

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
	/*	if err := host.getUserGroups(true); err != nil {
			return 0, err
		}

		host.VirtMu.Lock()
		defer host.VirtMu.Unlock()

		v := host.Virt

		usergroup := u.Group
		if usergroup == "" {
			usergroup = u.Name
		}

		userhome := u.Home
		if userhome == "" {
			userhome = "/home/" + u.Name
		}

		old := v.cacheUsers[u.Name]
		if host.Run.Dry {
			cachedold, hit := v.Users[u.Name]
			if hit {
				old = cachedold
			}
		}

		if u.Delete {
			if old == nil {
				return itemUnchanged, nil
			}
			if err := printExec(host, "userdel", old.Name); err != nil {
				return 0, err
			}
			v.Users[u.Name] = nil // not deleted on purpose: tombstone
			delete(v.cacheUsers, u.Name)
			return itemDeleted, nil
		}

		created := false
		modified := false

		if old == nil {
			//fmt.Printf("+ user %s (group %s)\n", u.Name, usergroup)
			created = true

			args := []string{"-d", userhome, "-g", usergroup, "-u", strconv.FormatUint(uint64(u.Uid), 10), u.Name}
			if u.Gecos != "" {
				args = append(args, "-c", u.Gecos)
			}
			if len(u.Groups) > 0 {
				args = append(args, "-G", strings.Join(u.Groups, ","))
			}
			if err := printExec(host, "useradd", args...); err != nil {
				return 0, err
			}
			old = &User{
				Name:   u.Name,
				Uid:    u.Uid,
				Group:  usergroup,
				Groups: u.Groups,
				Gecos:  u.Gecos,
			}
			v.Users[u.Name] = old
			v.cacheUsers[u.Name] = old
		}

		if old.Uid != u.Uid {
			//fmt.Printf("~ user %s (uid %d → %d)\n", u.Name, old.Uid, u.Uid)
			modified = true

			if err := printExec(host, "usermod", "-u", strconv.FormatUint(uint64(u.Uid), 10), u.Name); err != nil {
				return 0, err
			}
			old.Uid = u.Uid
		}

		oldpw := old.Password
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
		}

		resetgroups := false
		sort.Strings(old.Groups)
		sort.Strings(u.Groups)
		if len(old.Groups) != len(u.Groups) {
			resetgroups = true
		} else if len(u.Groups) > 0 {
			for i, gg := range old.Groups {
				if u.Groups[i] != gg {
					resetgroups = true
					break
				}
			}
		}
		if resetgroups {
			oldstr := strings.Join(old.Groups, ", ")
			newstr := strings.Join(u.Groups, ", ")
			if oldstr == "" {
				oldstr = "none"
			}
			if newstr == "" {
				newstr = "none"
			}
			modified = true
			//fmt.Printf("~ user %s groups (%s → %s)\n", u.Name, oldstr, newstr)
			if err := printExec(host, "usermod", "-G", strings.Join(u.Groups, ","), u.Name); err != nil {
				return 0, err
			}
			old.Groups = u.Groups
		}

		if old.Group != usergroup {
			modified = true
			//fmt.Printf("~ user %s (primary group %s → %s)\n", u.Name, old.Group, usergroup)
			if err := printExec(host, "usermod", "-g", usergroup, u.Name); err != nil {
				return 0, err
			}
			old.Group = usergroup
		}

		if old.Home != userhome {
			modified = true
			if err := printExec(host, "usermod", "-d", userhome, u.Name); err != nil {
				return 0, err
			}
			old.Home = userhome
		}

		if created {
			return itemCreated, nil
		}
		if modified {
			return itemModified, nil
		}*/
	return itemUnchanged, nil
}
