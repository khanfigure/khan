package local

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func (host *Host) Create(fpath string) (io.WriteCloser, error) {
	fmt.Println(host, ">", fpath)
	return os.Create(fpath)
}

func (host *Host) Remove(fpath string) error {
	fmt.Println(host, "$ rm", fpath)
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
	fmt.Printf("%s $ chmod %o %s\n", host, mode, fpath)
	return os.Chmod(fpath, mode)
}

func (host *Host) Chown(fpath string, uid uint32, gid uint32) error {
	fmt.Printf("%s $ chown %d:%d %s\n", host, uid, gid, fpath)
	return os.Chown(fpath, int(uid), int(gid))
}
