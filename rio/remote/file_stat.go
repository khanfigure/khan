package remote

import (
	"bytes"
	"os"
	"strings"

	"khan.rip/rio/util"

	"github.com/keegancsmith/shell"
)

func (host *Host) Stat(path string) (os.FileInfo, error) {
	// need this to know what args to pass to stat command
	info, err := host.Info()
	if err != nil {
		return nil, err
	}

	session, err := host.pool.Get(host.connect)
	if err != nil {
		return nil, err
	}
	defer session.Put()

	outbuf := &bytes.Buffer{}
	errbuf := &bytes.Buffer{}

	session.Stdout = outbuf
	session.Stderr = errbuf

	statcmd := "stat -t"
	if info.OS == "openbsd" {
		statcmd = "stat -r"
	}

	cmdline := statcmd + " " + shell.ReadableEscapeArg(path)

	err = session.Run(cmdline)

	outstr := strings.TrimSpace(outbuf.String())
	errstr := strings.TrimSpace(errbuf.String())

	return util.ParseStat(info.OS, path, outstr, errstr, err)
}
