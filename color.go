package duck

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/crypto/ssh/terminal"
)

type Color int

const (
	Black Color = iota + 1
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
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

func color(c Color) string {
	colorinit()
	if ttymode == -1 {
		return ""
	}
	return "\x1b[" + strconv.Itoa(int(c)+29) + "m"
}
func brightcolor(c Color) string {
	colorinit()
	if ttymode == -1 {
		return ""
	}
	return "\x1b[1;" + strconv.Itoa(int(c)+29) + "m"
}
func bg(c Color, b Color) string {
	colorinit()
	if ttymode == -1 {
		return ""
	}
	return "\x1b[" + strconv.Itoa(int(c)+29) + ";" + strconv.Itoa(int(b)+39) + "m"
}
func reset() string {
	colorinit()
	if ttymode == -1 {
		return ""
	}
	return "\x1b[m"
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
