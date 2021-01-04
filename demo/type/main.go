package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

const (
	first   = 300
	min     = 10
	max     = 100
	space   = 2
	newline = 0
)

// randomly slow version of io.copy
func slowcopy(dst io.Writer, src io.Reader) (written int64, err error) {
	buf := make([]byte, 1)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			slow := rand.Intn(max-min) + min
			if buf[0] == ' ' {
				slow *= space
			}
			if buf[0] == '\n' {
				slow *= newline
			}
			time.Sleep(time.Millisecond * time.Duration(slow))
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	return written, err
}

func main() {
	time.Sleep(time.Millisecond * first)
	_, err := slowcopy(os.Stdout, os.Stdin)
	if err != nil && err != io.EOF {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
