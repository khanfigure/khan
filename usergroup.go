package khan
/*
import (
	"bufio"
	"strconv"
	"strings"
)

// Always lock userCacheMu before calling this
func (r *Run) reloadUserGroupCache() error {
	// already cached!
	if r.groupCache != nil {
		return nil
	}

	r.groupCache = map[string]*Group{}
	r.gidCache = map[int]*Group{}
	r.userCache = map[string]*User{}
	r.uidCache = map[int]*User{}

	userGids := map[string]int{}

	u_rows, err := readColonFile(r, "/etc/passwd")
	if err != nil {
		return err
	}
	for _, row := range u_rows {
		if len(row) < 6 {
			continue
		}
		uid, err := strconv.Atoi(row[2])
		if err != nil {
			continue
		}
		gid, err := strconv.Atoi(row[3])
		if err != nil {
			continue
		}

		userGids[row[0]] = gid

		u := User{
			Name:  row[0],
			Uid:   uid,
			Gecos: row[4],
			Home:  row[5],
			Shell: row[6],
		}
		if len(u.Name) == 0 {
			continue
		}
		if u.Name[0] == '+' || u.Name[0] == '-' {
			continue
		}
		r.userCache[u.Name] = &u
		r.uidCache[u.Uid] = &u
	}

	sh_rows, err := readColonFile(r, "/etc/shadow")
	if err != nil && iserrnotfound(err) {
		// try openbsd mode!
		r.bsdmode = true
		sh_rows, err = readColonFile(r, "/etc/master.passwd")
	}
	if err != nil {
		return err
	}
	for _, row := range sh_rows {
		if len(row) < 8 {
			continue
		}
		u, ok := r.userCache[row[0]]
		if !ok {
			continue
		}

		if row[1] == "" {
			u.BlankPassword = true
		} else if row[1] == "!" || row[1] == "!!" || row[1] == "x" {
			// This is represented by:
			//    BlankPassword = false
			//    Password = ""
		} else {
			u.Password = row[1]
		}
		// TODO fancy /etc/shadow fields
	}

	g_rows, err := readColonFile(r, "/etc/group")
	if err != nil {
		return err
	}
	for _, row := range g_rows {
		if len(row) < 4 {
			continue
		}
		id, err := strconv.Atoi(row[2])
		if err != nil {
			continue
		}
		g := Group{
			Name: row[0],
			Gid:  id,
		}
		if len(g.Name) == 0 {
			continue
		}
		if g.Name[0] == '+' || g.Name[0] == '-' {
			continue
		}
		r.groupCache[g.Name] = &g
		r.gidCache[g.Gid] = &g
		for _, u := range strings.Split(row[3], ",") {
			u = strings.TrimSpace(u)
			uu, ok := r.userCache[u]
			if ok {
				uu.Groups = append(uu.Groups, g.Name)
			}
		}
	}

	for u, gid := range userGids {
		user := r.userCache[u]
		group := r.gidCache[gid]
		if user != nil && group != nil {
			user.Group = group.Name
		}
	}

	return nil
}

func readColonFile(r *Run, path string) ([][]string, error) {
	fh, err := r.rioconfig.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	var ret [][]string
	bs := bufio.NewScanner(fh)
	for bs.Scan() {
		line := bs.Text()
		comment := strings.IndexByte(line, '#')
		if comment != -1 {
			line = line[:comment]
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		vals := strings.Split(line, ":")
		for i, v := range vals {
			vals[i] = strings.TrimSpace(v)
		}
		ret = append(ret, vals)
	}
	return ret, bs.Err()
}*/
