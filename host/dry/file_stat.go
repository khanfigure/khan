package dry

import (
	"fmt"
	"os"
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
	return fmt.Sprintf("file %#v size %d mode %o mtime %v isdir %v uid %d gid %d",
		fi.name, fi.size, fi.mode, fi.modtime, fi.isdir, fi.uid, fi.gid)
}
