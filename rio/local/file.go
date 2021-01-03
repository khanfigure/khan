package local

import (
	"io"
	"io/ioutil"
	"os"
)

func (host *Host) Create(fpath string) (io.WriteCloser, error) {
	return os.Create(fpath)
}

func (host *Host) Remove(fpath string) error {
	return os.Remove(fpath)
}

func (host *Host) Open(fpath string) (io.ReadCloser, error) {
	return os.Open(fpath)
}

func (host *Host) ReadFile(fpath string) ([]byte, error) {
	return ioutil.ReadFile(fpath)
}

func (host *Host) Stat(fpath string) (os.FileInfo, error) {
	return os.Stat(fpath)
}

func (host *Host) Chmod(fpath string, mode os.FileMode) error {
	return os.Chmod(fpath, mode)
}

func (host *Host) Chown(fpath string, uid uint32, gid uint32) error {
	return os.Chown(fpath, int(uid), int(gid))
}
