package ksyn

import (
	"encoding/hex"
	"fmt"
)

type RGB struct {
	R, G, B uint8

	start, end Pos
}

func (rgb RGB) Start() Pos {
	return rgb.start
}

func (rgb RGB) End() Pos {
	return rgb.end
}

func (rgb RGB) FormatCSS() string {
	return fmt.Sprintf("#%02x%02x%02x", rgb.R, rgb.G, rgb.B)
}

func ch_is_rgb(ch byte) bool {
	if ch >= '0' && ch <= '9' ||
		ch >= 'a' && ch <= 'f' ||
		ch >= 'A' && ch <= 'F' {
		return true
	}
	return false
}

func (p *parser) rgb() (RGB, error) {
	if err := p.consumeRune('#'); err != nil {
		return RGB{}, err
	}

	s := p.str()

	i := 0

	ch := p.ch()
	for ch_is_rgb(ch) {
		i++
		p.next()
		ch = p.ch()
	}

	if i == 3 {
		s = string(s[0]) + string(s[0]) + string(s[1]) + string(s[1]) + string(s[2]) + string(s[2])
		v, err := hex.DecodeString(s)
		if err != nil {
			return RGB{}, err
		}
		return RGB{
			R: v[0],
			G: v[1],
			B: v[2],
		}, nil
	}
	if i == 6 {
		v, err := hex.DecodeString(s[:6])
		if err != nil {
			return RGB{}, err
		}
		return RGB{
			R: v[0],
			G: v[1],
			B: v[2],
		}, nil
	}

	return RGB{}, p.unexpected("3 or 6 hex characters")
}

func (rgb RGB) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T #%02x%02x%02x\n", rgb, rgb.R, rgb.G, rgb.B)
	return r
}
