package remote

import (
	"khan.rip/rio/util"
)

func (host *Host) MkdirAll(fpath string) error {
	return util.MkdirAll(host, fpath)
}
