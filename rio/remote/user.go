package remote

import (
	hhost "github.com/desops/khan/host"
	"github.com/desops/khan/host/util"
)

func (host *Host) Group(name string) (*hhost.Group, error) {
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

func (host *Host) CreateGroup(group *hhost.Group) error {
	if err := util.CreateGroup(host, group); err != nil {
		return err
	}

	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	if host.groups == nil {
		return nil
	}

	host.groups[group.Name] = group
	return nil
}

func (host *Host) UpdateGroup(group *hhost.Group) error {
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
