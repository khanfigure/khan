package util

import (
	"context"
	"os"
	"syscall"

	"khan.rip/rio"
)

func Remove(host rio.Host, fpath string) error {
	ctx := context.Background()
	if err := host.Exec(rio.Command(ctx, "rm", fpath)); err != nil {
		return err
	}
	return nil
}

func RemoveAll(host rio.Host, fpath string) error {
	ctx := context.Background()
	if err := host.Exec(rio.Command(ctx, "rm", "-rf", fpath)); err != nil {
		return err
	}
	return nil
}

func Rename(host rio.Host, oldpath, newpath string) error {
	ctx := context.Background()
	if err := host.Exec(rio.Command(ctx, "mv", oldpath, newpath)); err != nil {
		return err
	}
	return nil
}

func Mkdir(host rio.Host, fpath string) error {
	ctx := context.Background()
	if err := host.Exec(rio.Command(ctx, "mkdir", fpath)); err != nil {
		return err
	}
	return nil
}
func MkdirAll(host rio.Host, fpath string) error {
	ctx := context.Background()
	return host.Exec(rio.Command(ctx, "mkdir", "-p", fpath))
}

func IsErrNotFound(err error) bool {
	// TODO do this better
	v, ok := err.(*os.PathError)
	if ok && v != nil && v.Err == syscall.ENOENT {
		return true
	}
	return false
}
