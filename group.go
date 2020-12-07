package duck

type Group struct {
	Name string `duck:"name"`
	Gid  int    `duck:"gid"`

	id int
}

func (g *Group) setID(id int) {
	g.id = id
}
func (g *Group) getID() int {
	return g.id
}

func (g *Group) apply() error {
	return nil
}
