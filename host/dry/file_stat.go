package dry

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

type FileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modtime time.Time
	isdir   bool

	uid uint32
	gid uint32
}

func (fi *FileInfo) Name() string {
	return fi.name
}
func (fi *FileInfo) Size() int64 {
	return fi.size
}
func (fi *FileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi *FileInfo) ModTime() time.Time {
	return fi.modtime
}
func (fi *FileInfo) IsDir() bool {
	return fi.isdir
}
func (fi *FileInfo) Sys() interface{} {
	return fi
}
func (fi *FileInfo) Uid() uint32 {
	return fi.uid
}
func (fi *FileInfo) Gid() uint32 {
	return fi.gid
}
func (fi *FileInfo) String() string {
	return fmt.Sprintf("%T file %#v size %d mode %o mtime %v isdir %v uid %d gid %d",
		fi, fi.name, fi.size, fi.mode, fi.modtime, fi.isdir, fi.uid, fi.gid)
}

func (host *Host) Stat(fpath string) (os.FileInfo, error) {
	host.fsmu.Lock()
	file := host.fs[fpath]
	if file == nil && host.cascade != nil {
		// we don't want to hold this lock while SSH does its thing
		host.fsmu.Unlock()
		return host.cascade.Stat(fpath)
	}
	defer host.fsmu.Unlock()

	if file == nil || file.info == nil {
		return nil, &os.PathError{Op: "stat", Path: fpath, Err: syscall.ENOENT}
	}

	return file.info, nil
}
