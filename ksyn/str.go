package ksyn

import (
	"fmt"
)

type Str struct {
	Str string

	start, end Pos
}

func (s Str) Start() Pos {
	return s.start
}

func (s Str) End() Pos {
	return s.end
}

func (p *parser) exprStr() (Str, error) {
	start := p.pos

	if err := p.consumeRune('"'); err != nil {
		return Str{}, err
	}

	str := p.str()

	for i := 0; i < len(str); i++ {
		if str[i] == '"' {
			if err := p.consumeRune('"'); err != nil {
				return Str{}, err
			}
			end := p.pos
			return Str{
				Str:   str[:i],
				start: start,
				end:   end,
			}, nil
		}
		p.next()
	}

	return Str{}, p.unexpected("\"")
}

func (s Str) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T %#v\n", s, s.Str)
	return r
}
