package util

// Tools for emulating stat results (os.FileInfo implementation) and
// for parsing the output of the stat command on various operating systems

const (
	// Found these in OpenBSD's source for the stat command :)
	s_ifmt     = 0170000 // type of file mask
	s_ifdir    = 0040000 // directory
	s_ifreg    = 0100000 // regular
	s_justmode = 0777    // this is me ignoring things like suid for now
)

type FileInfo struct {
	Fname    string
	Fsize    int64
	Fmode    os.FileMode
	Fmodtime time.Time
	Fisdir   bool

	Fuid uint32
	Fgid uint32
}

func (fi *FileInfo) Name() string {
	return fi.Fname
}
func (fi *FileInfo) Size() int64 {
	return fi.Fsize
}
func (fi *FileInfo) Mode() os.FileMode {
	return fi.Fmode
}
func (fi *FileInfo) ModTime() time.Time {
	return fi.Fmodtime
}
func (fi *FileInfo) IsDir() bool {
	return fi.Fisdir
}
func (fi *FileInfo) Sys() interface{} {
	return fi
}
func (fi *FileInfo) Uid() uint32 {
	return fi.Fuid
}
func (fi *FileInfo) Gid() uint32 {
	return fi.Fgid
}

func (fi *FileInfo) String() string {
	return fmt.Sprintf("%T name %#v size %d mode %o mtime %v isdir %v uid %d gid %d",
		fi, fi.Fname, fi.Fsize, fi.Fmode, fi.Fmodtime, fi.Fisdir, fi.Fuid, fi.Fgid)
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

func ParseStat(osname, fpath, stdout, stderr string, execerr error) (*FileInfo, error) {
	if execerr != nil {
		if strings.HasPrefix(stderr, "stat: ") && strings.HasSuffix(stderr, "No such file or directory") {
			// emulate an OS stat call's error
			return nil, &os.PathError{
				Op:   "stat",
				Path: fpath,
				Err:  syscall.ENOENT,
			}
		}
		// some other execution error
		return nil, execerr
	}

	fi := &FileInfo{}

	switch osname {
	case "OpenBSD":
		match := openbsdStatRe.FindStringSubmatch(stdout)
		if match == nil {
			return nil, fmt.Errorf("Cannot parse OS %#v stat output: %#v", osname, outstr)
		}
		fi.Fname = match[6]
		if fi.Fsize, err = strconv.ParseInt(match[4], 10, 64); err != nil {
			return nil, err
		}
		mode, err := strconv.ParseUint(match[1], 8, 32)
		if err != nil {
			return nil, err
		}
		fi.Fmode = os.FileMode(mode)
		uid, err := strconv.ParseUint(match[2], 10, 32)
		if err != nil {
			return nil, err
		}
		fi.Fuid = uint32(uid)
		gid, err := strconv.ParseUint(match[3], 10, 32)
		if err != nil {
			return nil, err
		}
		fi.Fgid = uint32(gid)
		mtime, err := strconv.ParseInt(match[5], 10, 64)
		if err != nil {
			return nil, err
		}
		fi.Fmodtime = time.Unix(mtime, 0)

		if fi.Fmode&s_ifmt == s_ifdir {
			fi.Fisdir = true
		}
		return fi, nil

	case "Linux":
		match := linuxStatRe.FindStringSubmatch(outstr)
		if match == nil {
			return nil, fmt.Errorf("Cannot parse OS %#v stat output: %#v", osname, outstr)
		}
		fi.Fname = match[1]
		if fi.Fsize, err = strconv.ParseInt(match[2], 10, 64); err != nil {
			return nil, err
		}
		mode, err := strconv.ParseUint(match[3], 16, 32)
		if err != nil {
			return nil, err
		}
		fi.Fmode = os.FileMode(mode)
		uid, err := strconv.ParseUint(match[4], 10, 32)
		if err != nil {
			return nil, err
		}
		fi.Fuid = uint32(uid)
		gid, err := strconv.ParseUint(match[5], 10, 32)
		if err != nil {
			return nil, err
		}
		fi.Fgid = uint32(gid)
		mtime, err := strconv.ParseInt(match[6], 10, 64)
		if err != nil {
			return nil, err
		}
		fi.Fmodtime = time.Unix(mtime, 0)

		if fi.Fmode&s_ifmt == s_ifdir {
			fi.Fisdir = true
		}
		return fi, nil

	default:
		return nil, fmt.Errorf("Cannot parse stat output: Unhandled OS %#v", osname)
	}
}
