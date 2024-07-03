package khan

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

type Hue int

type Color struct {
	Bold            bool
	Dim             bool
	Italic          bool
	Underline       bool
	Strike          bool
	DoubleUnderline bool

	Color Hue
	Bg    Hue

	BrightColor bool
	BrightBg    bool
}

const (
	Unset Hue = iota
	Black
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	Grey
)

var ttymode int

func colorinit() {
	if ttymode == 0 {
		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			ttymode = 1
		} else {
			ttymode = -1
		}
	}
}

func (c Color) String() string {
	colorinit()
	if ttymode == -1 {
		return ""
	}

	r := "\x1b["
	var chunks []string
	if c.Bold {
		chunks = append(chunks, "1")
	}
	if c.Dim {
		chunks = append(chunks, "2")
	}
	if c.Italic {
		chunks = append(chunks, "3")
	}
	if c.Underline {
		chunks = append(chunks, "4")
	}
	if c.Strike {
		chunks = append(chunks, "9")
	}
	if c.DoubleUnderline {
		chunks = append(chunks, "21")
	}
	if c.Color != Unset {
		if c.BrightColor {
			chunks = append(chunks, strconv.Itoa(int(c.Color)+89))
		} else {
			chunks = append(chunks, strconv.Itoa(int(c.Color)+29))
		}
	}
	if c.Bg != Unset {
		if c.BrightBg {
			chunks = append(chunks, strconv.Itoa(int(c.Bg)+99))
		} else {
			chunks = append(chunks, strconv.Itoa(int(c.Bg)+39))
		}
	}
	r += strings.Join(chunks, ";") + "m"
	return r // + r[1:]
}
func (c Color) Wrap(s string) string {
	return c.String() + s + reset()
}

func color(c Hue) string {
	cc := Color{Color: c}
	return cc.String()
}
func boldcolor(c Hue) string {
	cc := Color{Color: c, Bold: true}
	return cc.String()
}
func reset() string {
	//	colorinit()
	//	if ttymode == -1 {
	//		return ""
	//	}
	//	return "\x1b[m"
	return Color{}.String()
}

func RedError(e error) string {
	return color(Red) + e.Error() + reset()
}

func RedPrintln(s string) {
	fmt.Println(color(Red) + s + reset())
}

func pass() string {
	return color(Green) + "✓" + reset()
}

func fail() string {
	return color(Red) + "✗" + reset()
}
