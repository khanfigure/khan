package khan

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

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

	pflag.StringVarP(&r.User, "user", "u", os.Getenv("USER"), "SSH User")

	var hostlist []string
	pflag.StringSliceVarP(&hostlist, "host", "h", nil, "SSH Host (may be host:port, may be repeated)")

	pflag.Parse()

	anyssh := false

	if localmode {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		r.Hosts = append(r.Hosts, &Host{
			Name: hostname,
			SSH:  false,
			Virt: NewVirtual(),
			Run:  r,
		})
	}

	if len(hostlist) > 0 {
		anyssh = true
		for _, h := range hostlist {
			name := h
			if i := strings.IndexByte(name, ':'); i > -1 {
				name = name[:i]
			}
			r.Hosts = append(r.Hosts, &Host{
				Name: name,
				SSH:  true,
				Host: h,
				Virt: NewVirtual(),
				Run:  r,
			})
		}
	}

	if len(r.Hosts) == 0 {
		fmt.Println("Nothing to do: No remote hosts (-h/--host) or local host (-l/--local) were specified")
		return nil
	}

	title := "███ "

	if r.Dry {
		title += "Dry running"
	} else {
		title += color(Green) + "Applying" + reset()
	}
	title += " " + brightcolor(Yellow) + r.title + reset() + " " + color(Yellow) + r.describe + reset() + " on "

	for i, host := range r.Hosts {
		if i > 0 {
			title += ", "
		}
		title += host.String()
	}
	fmt.Println(title)

	if anyssh {
		socket := os.Getenv("SSH_AUTH_SOCK")
		conn, err := net.Dial("unix", socket)
		if err != nil {
			return fmt.Errorf("Failed to open SSH_AUTH_SOCK: %w", err)
		}
		agentClient := agent.NewClient(conn)
		sshconfig := &ssh.ClientConfig{
			User: r.User,
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

	return r.run()
}
