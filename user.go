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

	// If UidPrimary (yaml: uid_primary) is true, the uid is treated as the primary identifier.
	// Behavior:
	//		UidPrimary true: usermod -l (name) is used if you change the name of the user
	//    UidPrimary false: usermod -u (uid) is used if you change the uid of the user
	UidPrimary bool `khan:"uid_primary"`

	Delete bool

	id int
}

func (u *User) String() string {
	return fmt.Sprintf("%s/%d", u.Name, u.Uid)
}

func (u *User) setID(id int) {
	u.id = id
}
func (u *User) getID() int {
	return u.id
}
func (u *User) apply(r *run) (itemStatus, error) {
	r.userCacheMu.Lock()
	defer r.userCacheMu.Unlock()

	if err := r.reloadUserGroupCache(); err != nil {
		return 0, err
	}

	usergroup := u.Group
	if usergroup == "" {
		usergroup = u.Name
	}

	var old *User

	if u.UidPrimary {
		old = r.uidCache[u.Uid]
	} else {
		old = r.userCache[u.Name]
	}

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

		args := []string{"-g", usergroup, "-u", strconv.Itoa(u.Uid), u.Name}
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
		input := bytes.NewBuffer([]byte(u.Name + ":" + newpw + "\n"))
		if err := printExecStdin(r, input, "chpasswd", "-e"); err != nil {
			return 0, err
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
	return itemUnchanged, nil
}
