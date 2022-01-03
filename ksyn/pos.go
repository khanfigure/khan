package ksyn

import (
	"fmt"
)

type Pos struct {
	Offset int
	Line   int
	Col    int

	Path string
}

func (p Pos) String() string {
	return fmt.Sprintf("%s:%d,%d", p.Path, p.Line+1, p.Col+1)
}
