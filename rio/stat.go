package rio

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/keegancsmith/shell"
)

const (
	// Found these in OpenBSD's source for the stat command :)
	s_ifmt  = 0170000
	s_ifdir = 0040000
)

type FileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modtime time.Time
	isdir   bool

	uid int
	gid int
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
	return nil
}

func (fi *FileInfo) String() string {
	return fmt.Sprintf("file %#v size %d mode %o mtime %v isdir %v uid %d gid %d",
		fi.name, fi.size, fi.mode, fi.modtime, fi.isdir, fi.uid, fi.gid)
}

// 10 17547654 drwxr-xr-x 2 joel joel 70100322 512 "Dec 18 19:52:23 2020" "Dec 18 19:52:23 2020" "Dec 18 19:52:23 2020" 32768 8 0 Hi There

// 10 17442839 0100644 1 1000 1000 0 0 1608352277 1608352277 1608352277 32768 0 0 dude

// st_dev, st_ino, st_mode, st_nlink, st_uid, st_gid, st_rdev, st_size, st_atime, st_mtime, st_ctime, st_blksize, st_blocks, st_flags, file name

// 1 mode (octal)
// 2 uid
// 3 gid
// 4 size
// 5 mtime
// 6 name
var openbsdStatRe = regexp.MustCompile(`^\d+ \d+ (\d+) \d+ (\d+) (\d+) \d+ (\d+) \d+ (\d+) \d+ \d+ \d+ \d+ (.*)$`)

//  File: file_heyo
//  Size: 2         	Blocks: 8          IO Block: 4096   regular file
//Device: 2dh/45d	Inode: 20619       Links: 1
//Access: (0644/-rw-r--r--)  Uid: ( 1000/    joel)   Gid: ( 1000/    joel)
//Access: 2020-12-18 19:51:25.706955391 -0700
//Modify: 2020-12-18 19:51:25.706955391 -0700
//Change: 2020-12-18 19:51:25.706955391 -0700
// Birth: -

// file_heyo 2 8 81a4 1000 1000 2d 20619 1 0 0 1608346285 1608346285 1608346285 0 4096
// %n %s %b %f %u %g %D %i %h %t %T %X %Y %Z %W %o

// 1 name
// 2 size
// 3 mode (hex)
// 4 uid
// 5 gid
// 6 mtime
var linuxStatRe = regexp.MustCompile(`^(.*) (\d+) \d+ ([a-fA-F0-9]+) (\d+) (\d+) [a-fA-F0-9]+ \d+ \d+ [a-fA-F0-9]+ [a-fA-F0-9]+ (\d+) \d+ \d+ \d+ \d+$`)

// /tmp/file_duck 9 8 81a4 1000 1000 2d 20963 1 0 0 1608356438 1608356438 1608356438 0 4096

func (config *Config) Stat(path string) (os.FileInfo, error) {
	if config.Pool == nil {
		return os.Stat(path)
	}

	// need this to know what args to pass to stat command
	ri, err := config.getremoteinfo(config.Host)
	if err != nil {
		return nil, err
	}

	statcmd := "stat -t"
	if ri.os == "OpenBSD" {
		statcmd = "stat -r"
	}

	session, err := config.Pool.Get(config.Host)
	if err != nil {
		return nil, err
	}
	defer session.Put()

	outbuf := &bytes.Buffer{}
	errbuf := &bytes.Buffer{}

	session.Stdout = outbuf
	session.Stderr = errbuf

	cmdline := statcmd + " " + shell.ReadableEscapeArg(path)
	if config.Sudo != "" {
		cmdline = "sudo -u " + shell.ReadableEscapeArg(config.Sudo) + " " + cmdline
	}

	if config.Verbose {
		fmt.Println("ssh", config.Host, cmdline)
	}

	if err := session.Run(cmdline); err != nil {
		if config.Verbose {
			fmt.Println("ssh", config.Host, cmdline, err)
		}

		e := strings.TrimSpace(errbuf.String())

		if strings.HasPrefix(e, "stat: ") && strings.HasSuffix(e, "No such file or directory") {
			// emulate os.Stat
			return nil, &os.PathError{
				Op:   "stat",
				Path: path,
				Err:  syscall.ENOENT,
			}
		}

		return nil, err
	}

	outstr := strings.TrimSpace(outbuf.String())

	fi := &FileInfo{}

	if ri.os == "OpenBSD" {
		match := openbsdStatRe.FindStringSubmatch(outstr)
		if match == nil {
			return nil, fmt.Errorf("Cannot parse OS %#v stat output: %#v", ri.os, outstr)
		}
		fi.name = match[6]
		if fi.size, err = strconv.ParseInt(match[4], 10, 64); err != nil {
			return nil, err
		}
		mode, err := strconv.ParseUint(match[1], 8, 32)
		if err != nil {
			return nil, err
		}
		fi.mode = os.FileMode(mode) // convert to uint32
		if fi.uid, err = strconv.Atoi(match[2]); err != nil {
			return nil, err
		}
		if fi.gid, err = strconv.Atoi(match[2]); err != nil {
			return nil, err
		}
		mtime, err := strconv.ParseInt(match[5], 10, 64)
		if err != nil {
			return nil, err
		}
		fi.modtime = time.Unix(mtime, 0)

		if fi.mode&s_ifmt == s_ifdir {
			fi.isdir = true
		}

		if config.Verbose {
			fmt.Println(fi)
		}

		return fi, nil
	}

	match := linuxStatRe.FindStringSubmatch(outstr)
	if match == nil {
		return nil, fmt.Errorf("Cannot parse OS %#v stat output: %#v", ri.os, outstr)
	}
	fi.name = match[1]
	if fi.size, err = strconv.ParseInt(match[2], 10, 64); err != nil {
		return nil, err
	}
	mode, err := strconv.ParseUint(match[3], 16, 32)
	if err != nil {
		return nil, err
	}
	fi.mode = os.FileMode(mode) // convert to uint32
	if fi.uid, err = strconv.Atoi(match[4]); err != nil {
		return nil, err
	}
	if fi.gid, err = strconv.Atoi(match[5]); err != nil {
		return nil, err
	}
	mtime, err := strconv.ParseInt(match[6], 10, 64)
	if err != nil {
		return nil, err
	}
	fi.modtime = time.Unix(mtime, 0)

	if fi.mode&s_ifmt == s_ifdir {
		fi.isdir = true
	}

	if config.Verbose {
		fmt.Println(fi)
	}

	return fi, nil
}
