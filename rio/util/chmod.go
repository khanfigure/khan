package util

import (
	"context"
	"fmt"
	"os"

	"github.com/desops/khan/rio"
)

func Chown(host rio.Host, fpath string, uid uint32, gid uint32) error {
	ctx := context.Background()
	if err := host.Exec(rio.Command(ctx, "chown", fmt.Sprintf("%d:%d", uid, gid), fpath)); err != nil {
		return err
	}
	return nil
}

func Chmod(host rio.Host, fpath string, perms os.FileMode) error {
	ctx := context.Background()
	if err := host.Exec(rio.Command(ctx, "chmod", fmt.Sprintf("%o", perms), fpath)); err != nil {
		return err
	}
	return nil
}
