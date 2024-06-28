package local

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

func (host *Host) Create(fpath string) (io.WriteCloser, error) {
	log.Println(host, ">", fpath)
	return os.Create(fpath)
}

func (host *Host) Remove(fpath string) error {
	log.Println(host, "! rm", fpath)
	return os.Remove(fpath)
}

func (host *Host) Rename(oldpath, newpath string) error {
	log.Println(host, "! mv", oldpath, newpath)
	return os.Rename(oldpath, newpath)
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
	log.Printf("%s ! chmod %o %s\n", host, mode, fpath)
	return os.Chmod(fpath, mode)
}

func (host *Host) Chown(fpath string, uid uint32, gid uint32) error {
	log.Printf("%s ! chown %d:%d %s\n", host, uid, gid, fpath)
	return os.Chown(fpath, int(uid), int(gid))
}
