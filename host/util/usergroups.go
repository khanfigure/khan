package util

import (
	"context"
	"strconv"
	"strings"

	hhost "github.com/desops/khan/host"
)

func LoadPasswords(host hhost.Host) (map[string]*hhost.Password, error) {
	info, err := host.Info()
	if err != nil {
		return nil, err
	}

	shadowfile := "/etc/shadow"
	if info.OS == "openbsd" {
		shadowfile = "/etc/master.passwd"
	}

	fh, err := host.Open(shadowfile)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	sh_rows, err := ParseColonFile(fh)
	if err != nil {
		return nil, err
	}

	r := map[string]*hhost.Password{}

	for _, row := range sh_rows {
		if len(row) < 8 {
			continue
		}
		r[row[0]] = &hhost.Password{
			Name:  row[0],
			Crypt: row[1],
		}
		// TODO fancy /etc/shadow fields for expiration and stuff
	}

	return r, nil
}

func LoadUserGroups(host hhost.Host) (map[string]*hhost.User, map[string]*hhost.Group, error) {
	//info, err := host.Info()
	//if err != nil {
	//	return nil, nil, err
	//}

	users := map[string]*hhost.User{}
	groups := map[string]*hhost.Group{}

	userGids := map[string]uint32{}
	gids := map[uint32]string{}

	fh, err := host.Open("/etc/passwd")
	if err != nil {
		return nil, nil, err
	}
	defer fh.Close()

	u_rows, err := ParseColonFile(fh)
	if err != nil {
		return nil, nil, err
	}
	for _, row := range u_rows {
		if len(row) < 6 {
			continue
		}
		uid, err := strconv.ParseUint(row[2], 10, 32)
		if err != nil {
			continue
		}
		gid, err := strconv.ParseUint(row[3], 10, 32)
		if err != nil {
			continue
		}

		userGids[row[0]] = uint32(gid)

		u := hhost.User{
			Name:  row[0],
			Uid:   uint32(uid),
			Home:  row[5],
			Shell: row[6],
		}
		if len(u.Name) == 0 {
			continue
		}
		if u.Name[0] == '+' || u.Name[0] == '-' {
			continue
		}
		users[u.Name] = &u
	}

	gfh, err := host.Open("/etc/group")
	if err != nil {
		return nil, nil, err
	}
	defer gfh.Close()

	g_rows, err := ParseColonFile(gfh)
	if err != nil {
		return nil, nil, err
	}
	for _, row := range g_rows {
		if len(row) < 4 {
			continue
		}
		id, err := strconv.ParseUint(row[2], 10, 32)
		if err != nil {
			continue
		}
		g := hhost.Group{
			Name: row[0],
			Gid:  uint32(id),
		}
		if len(g.Name) == 0 {
			continue
		}
		if g.Name[0] == '+' || g.Name[0] == '-' {
			continue
		}
		groups[g.Name] = &g
		gids[g.Gid] = g.Name
		for _, u := range strings.Split(row[3], ",") {
			u = strings.TrimSpace(u)
			uu, ok := users[u]
			if ok {
				uu.Groups = append(uu.Groups, g.Name)
			}
		}
	}

	for u, gid := range userGids {
		user := users[u]
		group := gids[gid]
		if user != nil && group != "" {
			user.Group = group
		}
	}

	return users, groups, nil
}

func CreateGroup(host hhost.Host, group *hhost.Group) error {
	ctx := context.Background()
	if err := host.Exec(hhost.Command(ctx, "groupadd", "-g", strconv.FormatUint(uint64(group.Gid), 10), group.Name)); err != nil {
		return err
	}
	return nil
}

func UpdateGroup(host hhost.Host, old *hhost.Group, group *hhost.Group) error {
	ctx := context.Background()
	if old.Gid != group.Gid {
		if err := host.Exec(hhost.Command(ctx, "groupmod", "-g", strconv.FormatUint(uint64(group.Gid), 10), group.Name)); err != nil {
			return err
		}
	}
	return nil
}

func DeleteGroup(host hhost.Host, name string) error {
	ctx := context.Background()
	if err := host.Exec(hhost.Command(ctx, "groupdel", name)); err != nil {
		return err
	}
	return nil
}
