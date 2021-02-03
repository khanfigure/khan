package dry

import (
	"fmt"

	"khan.rip/rio"
	"khan.rip/rio/util"
)

func (host *Host) Password(name string) (*rio.Password, error) {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	old, ok := host.passwords[name]
	if ok {
		return old, nil
	}
	if host.cascade != nil {
		return host.cascade.Password(name)
	}
	return nil, nil
}

func (host *Host) UpdatePassword(password *rio.Password) error {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	old, ok := host.passwords[password.Name]
	if !ok && host.cascade != nil {
		var err error
		old, err = host.cascade.Password(password.Name)
		if err != nil {
			return err
		}
	}
	if old == nil {
		return fmt.Errorf("Cannot set password: User %#v does not exist", password.Name)
	}
	if err := util.UpdatePassword(host, old, password); err != nil {
		return err
	}
	host.passwords[password.Name] = password
	return nil
}
