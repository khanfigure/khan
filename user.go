package khan

import (
	"fmt"
	/*"bytes"
	"sort"
	"strconv"
	"strings"*/)

type User struct {
	Name string

	// Primary group. If not specified, user name is used
	Group string

	Uid int

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
func (u *User) Needs() []string {
	return nil
}
func (u *User) Provides() []string {
	return nil
}

func (u *User) Apply(host *Host) (itemStatus, error) {
	return itemUnchanged, nil
	/*	r.userCacheMu.Lock()
		defer r.userCacheMu.Unlock()

		if err := r.reloadUserGroupCache(); err != nil {
			return 0, err
		}

		usergroup := u.Group
		if usergroup == "" {
			usergroup = u.Name
		}

		old := r.userCache[u.Name]

		if u.Delete {
			if old == nil {
				return itemUnchanged, nil
			}
			if err := printExec(r, "userdel", old.Name); err != nil {
				return 0, err
			}
			delete(r.userCache, old.Name)
			delete(r.uidCache, old.Uid)
			return itemDeleted, nil
		}

		created := false
		modified := false

		if old == nil {
			//fmt.Printf("+ user %s (group %s)\n", u.Name, usergroup)
			created = true

			args := []string{"-m", "-g", usergroup, "-u", strconv.Itoa(u.Uid), u.Name}
			if u.Gecos != "" {
				args = append(args, "-c", u.Gecos)
			}
			if len(u.Groups) > 0 {
				args = append(args, "-G", strings.Join(u.Groups, ","))
			}
			if err := printExec(r, "useradd", args...); err != nil {
				return 0, err
			}
			newuser := User{
				Name:   u.Name,
				Group:  usergroup,
				Groups: u.Groups,
				Gecos:  u.Gecos,
			}
			r.userCache[newuser.Name] = &newuser
			r.uidCache[newuser.Uid] = &newuser
		} else {
			if old.Name != u.Name {
				//fmt.Printf("~ uid %d (name %s → %s)\n", u.Uid, old.Name, u.Name)
				modified = true

				if err := printExec(r, "usermod", "-l", u.Name, old.Name); err != nil {
					return 0, err
				}
				newuser := *old
				newuser.Name = u.Name
				r.userCache[u.Name] = &newuser
				r.uidCache[u.Uid] = &newuser
				delete(r.userCache, old.Name)
			}
			if old.Uid != u.Uid {
				//fmt.Printf("~ user %s (uid %d → %d)\n", u.Name, old.Uid, u.Uid)
				modified = true

				if err := printExec(r, "usermod", "-u", strconv.Itoa(u.Uid), u.Name); err != nil {
					return 0, err
				}
				newuser := *old
				newuser.Uid = u.Uid
				r.userCache[u.Name] = &newuser
				r.uidCache[u.Uid] = &newuser
				delete(r.uidCache, old.Uid)
			}
		}

		old = r.userCache[u.Name]
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
			if r.bsdmode {
				// wish openbsd had chpasswd :'(
				// this leaks the crypted password hash via process args.
				// TODO maybe just buckle down and learn the proper way to lock the master file
				// and modify it directly?
				if err := printExec(r, "usermod", "-p", newpw, u.Name); err != nil {
					return 0, err
				}
			} else {
				input := bytes.NewBuffer([]byte(u.Name + ":" + newpw + "\n"))
				if err := printExecStdin(r, input, "chpasswd", "-e"); err != nil {
					return 0, err
				}
			}
			newuser := *old
			newuser.Password = u.Password
			newuser.BlankPassword = u.BlankPassword
			r.userCache[newuser.Name] = &newuser
			r.uidCache[newuser.Uid] = &newuser
		}

		old = r.userCache[u.Name]
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
			if err := printExec(r, "usermod", "-G", strings.Join(u.Groups, ","), u.Name); err != nil {
				return 0, err
			}
			old.Groups = u.Groups
		}

		old = r.userCache[u.Name]
		if old.Group != usergroup {
			modified = true
			//fmt.Printf("~ user %s (primary group %s → %s)\n", u.Name, old.Group, usergroup)
			if err := printExec(r, "usermod", "-g", usergroup, u.Name); err != nil {
				return 0, err
			}
			old.Group = usergroup
		}

		if created {
			return itemCreated, nil
		}
		if modified {
			return itemModified, nil
		}
		return itemUnchanged, nil*/
}
