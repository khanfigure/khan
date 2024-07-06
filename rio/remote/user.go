package remote

import (
	"khan.rip/rio"
	"khan.rip/rio/util"
)

func (host *Host) Group(name string) (*rio.Group, error) {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	if host.groups == nil {
		var err error
		host.users, host.groups, err = util.LoadUserGroups(host)
		if err != nil {
			return nil, err
		}
	}

	return host.groups[name], nil
}

func (host *Host) CreateGroup(group *rio.Group) error {
	gid, err := util.CreateGroup(host, group)
	if err != nil {
		return err
	}
	group.Gid = gid // Yuck

	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	if host.groups == nil {
		return nil
	}

	host.groups[group.Name] = group
	return nil
}

func (host *Host) UpdateGroup(group *rio.Group) error {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	if host.groups == nil {
		var err error
		host.users, host.groups, err = util.LoadUserGroups(host)
		if err != nil {
			return err
		}
	}

	old := host.groups[group.Name]

	if err := util.UpdateGroup(host, old, group); err != nil {
		return err
	}

	host.groups[group.Name] = group
	return nil
}

func (host *Host) DeleteGroup(name string) error {
	if err := util.DeleteGroup(host, name); err != nil {
		return err
	}

	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	if host.groups == nil {
		return nil
	}

	delete(host.groups, name)
	return nil
}

func (host *Host) User(name string) (*rio.User, error) {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	if host.users == nil {
		var err error
		host.users, host.groups, err = util.LoadUserGroups(host)
		if err != nil {
			return nil, err
		}
	}

	return host.users[name], nil
}

func (host *Host) CreateUser(user *rio.User) error {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	uid, err := util.CreateUser(host, user)
	if err != nil {
		return err
	}

	user.Uid = uid // Yuck
	host.users[user.Name] = user
	return nil
}

func (host *Host) UpdateUser(user *rio.User) error {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	old := host.users[user.Name]

	if err := util.UpdateUser(host, old, user); err != nil {
		return err
	}

	host.users[user.Name] = user
	return nil
}

func (host *Host) DeleteUser(name string) error {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	if err := util.DeleteUser(host, name); err != nil {
		return err
	}

	delete(host.users, name)
	return nil
}
