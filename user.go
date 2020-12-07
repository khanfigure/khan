package duck

type User struct {
	Name     string `duck:"name"`
	Password string `duck:"password"`
	Group    string `duck:"group"`

	Uid    int      `duck:"uid"`
	Gid    int      `duck:"gid"`
	Groups []string `duck:"groups"`

	id int
}

func (u *User) setID(id int) {
	u.id = id
}
func (u *User) getID() int {
	return u.id
}
func (u *User) apply() error {
	return nil
}
