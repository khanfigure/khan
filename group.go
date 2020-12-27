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

func (g *Group) setID(id int) {
	g.id = id
}
func (g *Group) getID() int {
	return g.id
}
func (g *Group) needs() []string {
	return nil
}
func (g *Group) provides() []string {
	return nil
}

func (g *Group) apply(r *Run) (itemStatus, error) {
	r.userCacheMu.Lock()
	defer r.userCacheMu.Unlock()

	if err := r.reloadUserGroupCache(); err != nil {
		return 0, err
	}

	old := r.groupCache[g.Name]

	if g.Delete {
		if old == nil {
			return itemUnchanged, nil
		}
		if err := printExec(r, "groupdel", old.Name); err != nil {
			return 0, err
		}
		delete(r.groupCache, old.Name)
		delete(r.gidCache, old.Gid)
		return itemDeleted, nil
	}

	if old == nil {
		//fmt.Printf("+ group %s (gid %d)\n", g.Name, g.Gid)
		if err := printExec(r, "groupadd", "-g", strconv.Itoa(g.Gid), g.Name); err != nil {
			return 0, err
		}
		newgrp := Group{
			Name: g.Name,
			Gid:  g.Gid,
		}
		r.groupCache[g.Name] = &newgrp
		r.gidCache[g.Gid] = &newgrp
		return itemCreated, nil
	}

	modified := false

	if old.Name != g.Name {
		modified = true
		//fmt.Printf("~ gid %d (name %s → %s)\n", g.Gid, old.Name, g.Name)
		if err := printExec(r, "groupmod", "-n", g.Name, old.Name); err != nil {
			return 0, err
		}
		newgrp := Group{
			Name: g.Name,
			Gid:  g.Gid,
		}
		r.groupCache[newgrp.Name] = &newgrp
		r.gidCache[newgrp.Gid] = &newgrp
		delete(r.groupCache, old.Name)

		for _, u := range r.userCache {
			for i, ug := range u.Groups {
				if ug == old.Name {
					u.Groups[i] = g.Name
				}
			}
			if u.Group == old.Name {
				u.Group = g.Name
			}
		}
	} else if old.Gid != g.Gid {
		modified = true
		//fmt.Printf("~ group %s (gid %d → %d)\n", g.Name, old.Gid, g.Gid)
		if err := printExec(r, "groupmod", "-g", strconv.Itoa(g.Gid), g.Name); err != nil {
			return 0, err
		}
		newgrp := Group{
			Name: g.Name,
			Gid:  g.Gid,
		}
		r.groupCache[g.Name] = &newgrp
		r.gidCache[g.Gid] = &newgrp
		delete(r.gidCache, old.Gid)
	}

	if modified {
		return itemModified, nil
	}

	return itemUnchanged, nil
}
