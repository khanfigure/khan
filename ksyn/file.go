package ksyn

import (
	"fmt"
	"io/ioutil"
)

type File struct {
	Path  string
	Stmts []Stmt

	start, end Pos
}

func (f File) Start() Pos {
	return f.start
}

func (f File) End() Pos {
	return f.end
}

func ParseFile(filepath string) (File, error) {
	buf, err := ioutil.ReadFile(filepath)
	if err != nil {
		return File{}, err
	}

	source := string(buf)

	p := parser{
		pos: Pos{
			Path: filepath,
		},
		source: source,
	}

	start := p.pos

	if _, err := p.comments(); err != nil {
		return File{}, err
	}

	stmts, err := p.stmts()
	if err != nil {
		return File{}, err
	}

	if _, err := p.comments(); err != nil {
		return File{}, err
	}

	if p.ch() != 0 {
		dangling := p.str()
		if len(dangling) > 64 {
			dangling = dangling[:64] + " ... (trimmed)"
		}
		return File{}, fmt.Errorf("Unparsed input dangling: %#v", dangling)
	}

	end := p.pos

	return File{
		Path:  filepath,
		Stmts: stmts,

		start: start,
		end:   end,
	}, nil
}

func (f File) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T %#v\n", f, f.Path)
	r += reprStmts(f.Stmts, style, pre1, pre2)
	return r
}
