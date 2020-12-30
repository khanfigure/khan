package local

type Host struct {
}

func (host *Host) String() string {
	return "local"
}

func New() *Host {
	return &Host{}
}
