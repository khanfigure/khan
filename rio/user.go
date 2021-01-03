package rio

type User struct {
	Name    string
	Uid     uint32
	Group   string
	Groups  []string
	Home    string
	Shell   string
	Comment string
}

type Password struct {
	Name  string
	Crypt string
}

type Group struct {
	Name string
	Gid  uint32
}
