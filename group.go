package khan

import (
	"fmt"
	"strconv"
)

type Group struct {
	Name string
	Gid  uint32

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
func (g *Group) After() []string {
	return nil
}
func (g *Group) Before() []string {
	return nil
}
func (g *Group) Provides() []string {
	if g.Delete {
		return []string{"-group:" + g.Name}
	} else {
		return []string{"group:" + g.Name}
	}
}

func (g *Group) Apply(host *Host) (itemStatus, error) {
	if err := host.getUserGroups(true); err != nil {
		return 0, err
	}

	host.VirtMu.Lock()
	defer host.VirtMu.Unlock()

	v := host.Virt

	old := v.cacheGroups[g.Name]
	if host.Run.Dry {
		cachedold, hit := v.Groups[g.Name]
		if hit {
			old = cachedold
		}
	}

	if g.Delete {
		if old == nil {
			return itemUnchanged, nil
		}
		if err := printExec(host, "groupdel", g.Name); err != nil {
			return 0, err
		}
		v.Groups[g.Name] = nil // not deleted on purpose: tombstone
		delete(v.cacheGroups, old.Name)
		return itemDeleted, nil
	}

	if old == nil {
		//fmt.Printf("+ group %s (gid %d)\n", g.Name, g.Gid)
		if err := printExec(host, "groupadd", "-g", strconv.FormatUint(uint64(g.Gid), 10), g.Name); err != nil {
			return 0, err
		}
		old = &Group{
			Name: g.Name,
			Gid:  g.Gid,
		}
		v.Groups[g.Name] = old
		v.cacheGroups[g.Name] = old
		return itemCreated, nil
	}

	modified := false

	if old.Gid != g.Gid {
		modified = true
		//fmt.Printf("~ group %s (gid %d → %d)\n", g.Name, old.Gid, g.Gid)
		if err := printExec(host, "groupmod", "-g", strconv.FormatUint(uint64(g.Gid), 10), g.Name); err != nil {
			return 0, err
		}
		old.Gid = g.Gid
	}

	if modified {
		return itemModified, nil
	}

	return itemUnchanged, nil
}
