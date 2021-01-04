package khan

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"khan.rip/rio"
	"khan.rip/rio/dry"
	"khan.rip/rio/local"
	"khan.rip/rio/remote"

	"github.com/desops/sshpool"
	"github.com/flosch/pongo2/v4"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var (
	defaultrun *Run = &Run{
		meta:    map[int]*imeta{},
		fences:  map[string]*sync.Mutex{},
		befores: map[string][]string{},
		errors:  map[string]error{},
	}
)

func SetSourcePrefix(s string) {
	defaultrun.sourceprefix = s
}
func SetDescribe(s string) {
	defaultrun.describe = s
}
func SetTitle(s string) {
	defaultrun.title = s
}

func Apply() error {
	r := defaultrun

	r.assetfn = mainassetfn

	r.pongocachefiles = map[string]*pongo2.Template{}
	r.pongocachestrings = map[string]*pongo2.Template{}
	r.pongopackedset = pongo2.NewSet("packed", &bindataloader{r})
	r.pongopackedcontext = pongo2.Context{
		"khan": map[string]interface{}{},
	}

	//r.rioconfig = &rio.Config{}

	pflag.BoolVarP(&r.Dry, "dry", "d", false, "Dry run; Don't make any changes")
	pflag.BoolVarP(&r.Diff, "diff", "D", false, "Show full diff of file content changes")
	pflag.BoolVarP(&r.Verbose, "verbose", "v", false, "Be more verbose")

	localmode := false
	pflag.BoolVarP(&localmode, "local", "l", false, "Run without SSH against local host as current user")

	var hostlist []string
	pflag.StringSliceVarP(&hostlist, "remote", "r", nil, "Run against remote host via SSH (user@host:port, may be repeated)")

	pflag.Parse()

	if localmode {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		rh := rio.Host(local.New())
		if r.Dry {
			rh = rio.Host(dry.New(uint32(os.Geteuid()), uint32(os.Getegid()), rh))
		}

		r.Hosts = append(r.Hosts, &Host{
			Name: hostname,
			SSH:  false,
			Run:  r,
			rh:   rh,
		})
	}

	for _, h := range hostlist {
		if r.Pool == nil {
			// initialize SSH pool
			socket := os.Getenv("SSH_AUTH_SOCK")
			conn, err := net.Dial("unix", socket)
			if err != nil {
				return fmt.Errorf("Failed to open SSH_AUTH_SOCK: %w", err)
			}
			agentClient := agent.NewClient(conn)
			sshconfig := &ssh.ClientConfig{
				Auth: []ssh.AuthMethod{
					ssh.PublicKeysCallback(agentClient.Signers),
				},
				HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
					// TODO
					return nil
				},
				BannerCallback: ssh.BannerDisplayStderr(),
			}

			r.Pool = sshpool.New(sshconfig, &sshpool.PoolConfig{Debug: false}) //r.Verbose})
		}
		name := h
		if i := strings.IndexByte(name, ':'); i > -1 {
			name = name[:i]
		}

		rh := rio.Host(remote.New(r.Pool, h))
		if r.Dry {
			// This uid/gid guess is incorrect. TODO: Concurrently SSH to all the hosts and
			// get this info correctly. This could double-serve as a pool warmup :)
			uid := os.Geteuid()
			gid := os.Getegid()
			at := strings.IndexByte(h, '@')
			if at > -1 && h[:at] == "root" {
				uid = 0
				gid = 0
			}
			rh = rio.Host(dry.New(uint32(uid), uint32(gid), rh))
		}

		r.Hosts = append(r.Hosts, &Host{
			Name: name,
			SSH:  true,
			Host: h,
			Run:  r,
			rh:   rh,
		})
	}

	if len(r.Hosts) == 0 {
		fmt.Println("Nothing to do: No remote hosts (-h/--host) or local host (-l/--local) were specified")
		return nil
	}

	decorate := color(Cyan) + "░▒▓█" + reset()

	title := decorate + " "

	if r.Dry {
		title += "Dry running"
	} else {
		title += "Applying"
	}
	title += " " + brightcolor(Yellow) + r.title + reset()
	if r.describe != "" && r.describe != "unknown" {
		title += " " + color(Yellow) + r.describe + reset()
	}
	title += " on "

	for i, host := range r.Hosts {
		if i > 0 {
			title += ", "
		}
		title += host.String()
	}
	fmt.Println(title)

	if err := r.run(); err != nil {
		return err
	}

	fmt.Println(decorate + " " + color(Green) + "✓" + reset() + " Great success!")
	return nil
}
