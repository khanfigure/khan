package khan

import (
	"fmt"
	"strconv"
)

type Group struct {
	Name string
	Gid  int

	Delete bool

	id int
}

func (g *Group) String() string {
	return fmt.Sprintf("%s/%d", g.Name, g.Gid)
}

func (g *Group) SetID(id int) {
	g.id = id
}
func (g *Group) ID() int {
	return g.id
}
func (g *Group) Clone() Item {
	r := *g
	r.id = 0
	return &r
}
func (g *Group) Needs() []string {
	return nil
}
func (g *Group) Provides() []string {
	return nil
}

func (g *Group) Apply(host *Host) (itemStatus, error) {
	if err := host.getUserGroups(); err != nil {
		return 0, err
	}

	host.VirtMu.Lock()
	defer host.VirtMu.Unlock()

	v := host.Virt

	old := v.Groups[g.Name]

	if g.Delete {
		if old == nil {
			return itemUnchanged, nil
		}
		if err := printExec(host, "groupdel", old.Name); err != nil {
			return 0, err
		}
		delete(v.Groups, old.Name)
		return itemDeleted, nil
	}

	if old == nil {
		//fmt.Printf("+ group %s (gid %d)\n", g.Name, g.Gid)
		if err := printExec(host, "groupadd", "-g", strconv.Itoa(g.Gid), g.Name); err != nil {
			return 0, err
		}
		newgrp := Group{
			Name: g.Name,
			Gid:  g.Gid,
		}
		v.Groups[g.Name] = &newgrp
		return itemCreated, nil
	}

	modified := false

	if old.Gid != g.Gid {
		modified = true
		//fmt.Printf("~ group %s (gid %d → %d)\n", g.Name, old.Gid, g.Gid)
		if err := printExec(host, "groupmod", "-g", strconv.Itoa(g.Gid), g.Name); err != nil {
			return 0, err
		}
		v.Groups[g.Name].Gid = g.Gid
	}

	if modified {
		return itemModified, nil
	}

	return itemUnchanged, nil
}
