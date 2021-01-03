package dry

import (
	"fmt"

	hhost "github.com/desops/khan/host"
	"github.com/desops/khan/host/util"
)

func (host *Host) Group(name string) (*hhost.Group, error) {
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

func (host *Host) CreateGroup(group *hhost.Group) error {
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

	if err := util.CreateGroup(host, group); err != nil {
		return err
	}
	host.groups[group.Name] = group
	return nil
}

func (host *Host) UpdateGroup(group *hhost.Group) error {
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
