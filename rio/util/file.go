package util

import (
	"context"

	"github.com/desops/khan/rio"
)

func Remove(host rio.Host, fpath string) error {
	ctx := context.Background()
	if err := host.Exec(rio.Command(ctx, "rm", fpath)); err != nil {
		return err
	}
	return nil
}
