package ksyn

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	files, err := filepath.Glob("tests/*.b")
	if err != nil {
		t.Error(err)
		return
	}
	for _, filename := range files {
		t.Run(filename, func(t *testing.T) {
			inf := filename
			outf := filename + ".out"
			errf := filename + ".err"
			node, err := ParseFile(inf)
			if err != nil {
				errbuf, errerr := ioutil.ReadFile(errf)
				if errerr != nil {
					t.Error(err)
					t.Error(errerr)
					return
				}
				errbuf_s := strings.TrimSpace(string(errbuf))
				err_s := strings.TrimSpace(err.Error())
				if !strings.Contains(err_s, errbuf_s) {
					t.Error(fmt.Errorf("Expected error substring %#v (from %#v): Got %#v", errbuf_s, errf, err_s))
				}
				return
			}
			have := node.repr(minimalReprStyle, "", "")
			out, err := ioutil.ReadFile(outf)
			if err != nil {
				t.Error(err)
				return
			}
			want := strings.TrimSpace(string(out))
			if want != strings.TrimSpace(have) {
				fmt.Println(have)
				t.Error(fmt.Errorf("Output mismatch from %#v to %#v", inf, outf))
			}
		})
	}
}
