package khan

import (
	"fmt"

	"khan.rip/rio"
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

func (g *Group) Apply(host *Host) (Status, error) {
	old, err := host.rh.Group(g.Name)
	if err != nil {
		return 0, err
	}

	if g.Delete {
		if old == nil {
			return Unchanged, nil
		}
		if err := host.rh.DeleteGroup(g.Name); err != nil {
			return 0, err
		}
		return Deleted, nil
	}

	v := &rio.Group{
		Name: g.Name,
		Gid:  g.Gid,
	}

	if old == nil {
		if err := host.rh.CreateGroup(v); err != nil {
			return 0, err
		}
		return Created, nil
	}

	if old.Gid != g.Gid {
		if err := host.rh.UpdateGroup(v); err != nil {
			return 0, err
		}
		return Modified, nil
	}

	return Unchanged, nil
}
