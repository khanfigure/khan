package util

import (
	"bufio"
	"io"
	"strings"
)

// ParseColonFile parses a file in the style of /etc/passwd or /etc/shadow.
func ParseColonFile(r io.Reader) ([][]string, error) {
	var ret [][]string
	bs := bufio.NewScanner(r)
	for bs.Scan() {
		line := bs.Text()
		comment := strings.IndexByte(line, '#')
		if comment != -1 {
			line = line[:comment]
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		vals := strings.Split(line, ":")
		for i, v := range vals {
			vals[i] = strings.TrimSpace(v)
		}
		ret = append(ret, vals)
	}
	return ret, bs.Err()
}
