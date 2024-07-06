package dry

import (
	"fmt"

	"khan.rip/rio"
	"khan.rip/rio/util"
)

func (host *Host) Group(name string) (*rio.Group, error) {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	old, ok := host.groups[name]
	if ok {
		return old, nil
	}
	if host.cascade != nil {
		return host.cascade.Group(name)
	}
	return nil, nil
}

func (host *Host) CreateGroup(group *rio.Group) error {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	old, ok := host.groups[group.Name]
	if !ok && host.cascade != nil {
		var err error
		old, err = host.cascade.Group(group.Name)
		if err != nil {
			return err
		}
	}
	if old != nil {
		return fmt.Errorf("Group %#v already exists", group.Name)
	}

	gid, err := util.CreateGroup(host, group)
	if err != nil {
		return err
	}
	group.Gid = gid // Yuck
	host.groups[group.Name] = group
	return nil
}

func (host *Host) UpdateGroup(group *rio.Group) error {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	old, ok := host.groups[group.Name]
	if !ok && host.cascade != nil {
		var err error
		old, err = host.cascade.Group(group.Name)
		if err != nil {
			return err
		}
	}
	if old == nil {
		return fmt.Errorf("Group %#v does not exist", group.Name)
	}
	if err := util.UpdateGroup(host, old, group); err != nil {
		return err
	}
	host.groups[group.Name] = group
	return nil
}

func (host *Host) DeleteGroup(name string) error {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	old, ok := host.groups[name]
	if !ok && host.cascade != nil {
		var err error
		old, err = host.cascade.Group(name)
		if err != nil {
			return err
		}
	}
	if old == nil {
		return fmt.Errorf("Group %#v does not exist", name)
	}
	if err := util.DeleteGroup(host, name); err != nil {
		return err
	}
	host.groups[name] = nil // tombstone
	return nil
}

func (host *Host) User(name string) (*rio.User, error) {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	old, ok := host.users[name]
	if ok {
		return old, nil
	}
	if host.cascade != nil {
		return host.cascade.User(name)
	}
	return nil, nil
}

func (host *Host) CreateUser(user *rio.User) error {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	old, ok := host.users[user.Name]
	if !ok && host.cascade != nil {
		var err error
		old, err = host.cascade.User(user.Name)
		if err != nil {
			return err
		}
	}
	if old != nil {
		return fmt.Errorf("User %#v already exists", user.Name)
	}

	uid, err := util.CreateUser(host, user)
	if err != nil {
		return err
	}
	user.Uid = uid // Yuck
	host.users[user.Name] = user
	host.passwords[user.Name] = &rio.Password{
		Name:  user.Name,
		Crypt: "!", // maybe be fancy later and make "*" if cascade upstream is openbsd
	}

	return nil
}

func (host *Host) UpdateUser(user *rio.User) error {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	old, ok := host.users[user.Name]
	if !ok && host.cascade != nil {
		var err error
		old, err = host.cascade.User(user.Name)
		if err != nil {
			return err
		}
	}
	if old == nil {
		return fmt.Errorf("User %#v does not exist", user.Name)
	}
	if err := util.UpdateUser(host, old, user); err != nil {
		return err
	}
	host.users[user.Name] = user
	return nil
}

func (host *Host) DeleteUser(name string) error {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	old, ok := host.users[name]
	if !ok && host.cascade != nil {
		var err error
		old, err = host.cascade.User(name)
		if err != nil {
			return err
		}
	}
	if old == nil {
		return fmt.Errorf("User %#v does not exist", name)
	}
	if err := util.DeleteUser(host, name); err != nil {
		return err
	}
	host.users[name] = nil // tombstone
	return nil
}
