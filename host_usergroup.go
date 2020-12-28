package khan

import (
	"bufio"
	"strconv"
	"strings"
)

// Always lock userCacheMu before calling this
func (host *Host) getUserGroups() error {
	if err := host.getInfo(); err != nil {
		return err
	}

	host.VirtMu.Lock()
	defer host.VirtMu.Unlock()

	v := host.Virt

	if v.Users != nil {
		// already cached!
		return nil
	}

	v.Users = map[string]*User{}
	v.Groups = map[string]*Group{}

	userGids := map[string]int{}
	gids := map[int]string{}

	u_rows, err := readColonFile(host, "/etc/passwd")
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
		v.Users[u.Name] = &u
	}

	shadowfile := "/etc/shadow"
	if v.OS == "OpenBSD" {
		shadowfile = "/etc/master.passwd"
	}

	sh_rows, err := readColonFile(host, shadowfile)
	if err != nil {
		return err
	}
	for _, row := range sh_rows {
		if len(row) < 8 {
			continue
		}
		u, ok := v.Users[row[0]]
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

	g_rows, err := readColonFile(host, "/etc/group")
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
		v.Groups[g.Name] = &g
		gids[g.Gid] = g.Name
		for _, u := range strings.Split(row[3], ",") {
			u = strings.TrimSpace(u)
			uu, ok := v.Users[u]
			if ok {
				uu.Groups = append(uu.Groups, g.Name)
			}
		}
	}

	for u, gid := range userGids {
		user := v.Users[u]
		group := gids[gid]
		if user != nil && group != "" {
			user.Group = group
		}
	}

	return nil
}

func readColonFile(host *Host, path string) ([][]string, error) {
	fh, err := host.Open(path)
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
}
