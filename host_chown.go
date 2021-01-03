package khan

/*
import (
	"context"
	"fmt"
	"os"
)

func (host *Host) Chown(path string, uid uint32, gid uint32) error {
	if !host.SSH {
		return os.Chown(path, int(uid), int(gid))
	}

	cmd := host.Command(context.Background(), "chown", fmt.Sprintf("%d:%d", uid, gid), path)

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (host *Host) Chmod(path string, perms os.FileMode) error {
	if !host.SSH {
		return os.Chmod(path, perms)
	}

	cmd := host.Command(context.Background(), "chmod", fmt.Sprintf("%o", perms), path)

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}*/
