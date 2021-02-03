package remote

import (
	"khan.rip/rio"
	"khan.rip/rio/util"
)

func (host *Host) Password(name string) (*rio.Password, error) {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	if host.passwords == nil {
		var err error
		host.passwords, err = util.LoadPasswords(host)
		if err != nil {
			return nil, err
		}
	}

	return host.passwords[name], nil
}

func (host *Host) UpdatePassword(password *rio.Password) error {
	host.usersmu.Lock()
	defer host.usersmu.Unlock()

	old := host.passwords[password.Name]

	if err := util.UpdatePassword(host, old, password); err != nil {
		return err
	}

	host.passwords[password.Name] = password
	return nil
}
