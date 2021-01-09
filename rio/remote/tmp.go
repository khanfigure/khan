package remote

import (
	"bytes"
	"context"
	"strings"

	"khan.rip/rio"
	"khan.rip/rio/util"
)

func (host *Host) TmpFile() (string, error) {
	tmpdir, err := host.TmpDir()
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	cmd := rio.Command(ctx, "mktemp", "-p", tmpdir, "XXXXXXXX")

	buf := &bytes.Buffer{}
	cmd.Stdout = buf

	if err := host.Exec(cmd); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

func (host *Host) TmpDir() (string, error) {
	host.tmpdirmu.Lock()
	defer host.tmpdirmu.Unlock()

	if host.tmpdir != "" {
		return host.tmpdir, nil
	}

	ctx := context.Background()
	cmd := rio.Command(ctx, "mktemp", "-d", "/tmp/tmpkhan_XXXXXXXX")

	buf := &bytes.Buffer{}
	cmd.Stdout = buf

	if err := host.Exec(cmd); err != nil {
		return "", err
	}

	fpath := strings.TrimSpace(buf.String())
	host.tmpdir = fpath
	return fpath, nil
}

func (host *Host) Cleanup() error {
	host.tmpdirmu.Lock()
	defer host.tmpdirmu.Unlock()

	if host.tmpdir == "" {
		return nil
	}
	if err := util.RemoveAll(host, host.tmpdir); err != nil {
		return err
	}
	return nil
}
