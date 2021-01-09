package local

import (
	"fmt"
	"io/ioutil"
	"os"
)

func (host *Host) TmpFile() (string, error) {
	tmpdir, err := host.TmpDir()
	if err != nil {
		return "", err
	}

	fmt.Println(host, "! mktemp -p", tmpdir, "XXXXXXXX")

	f, err := ioutil.TempFile(tmpdir, "")
	if err != nil {
		return "", err
	}
	defer f.Close()

	return f.Name(), nil
}

func (host *Host) TmpDir() (string, error) {
	host.tmpdirmu.Lock()
	defer host.tmpdirmu.Unlock()

	if host.tmpdir != "" {
		return host.tmpdir, nil
	}

	fmt.Println(host, "! mktemp -d /tmp/tmpkhan_XXXXXXXX")

	fpath, err := ioutil.TempDir("", "tmpkhan_")
	if err != nil {
		return "", err
	}
	host.tmpdir = fpath
	return fpath, nil
}

func (host *Host) Cleanup() error {
	host.tmpdirmu.Lock()
	defer host.tmpdirmu.Unlock()

	if host.tmpdir == "" {
		return nil
	}
	fmt.Println(host, "! rm -rf", host.tmpdir)
	if err := os.RemoveAll(host.tmpdir); err != nil {
		return err
	}
	return nil
}
