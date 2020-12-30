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
