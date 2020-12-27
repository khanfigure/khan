package khan

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/desops/khan/rio"

	"github.com/desops/sshpool"
	"github.com/flosch/pongo2/v4"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var (
	describe     string
	sourceprefix string

	run *Run = &Run{}
)

func SetSourcePrefix(s string) {
	sourceprefix = s
}
func SetDescribe(s string) {
	describe = s
}

func Apply() error {

	r := run

	r.assetfn = mainassetfn

	r.pongocachefiles = map[string]*pongo2.Template{}
	r.pongocachestrings = map[string]*pongo2.Template{}
	r.pongopackedset = pongo2.NewSet("packed", &bindataloader{r})
	r.pongopackedcontext = pongo2.Context{
		"khan": map[string]interface{}{
			"secret": func(path string) (map[string]string, error) {
				buf := &bytes.Buffer{}
				cmd := r.rioconfig.Command(context.Background(), "vault", "kv", "get", "-format", "json", "secret/"+path)
				cmd.Shell = true
				cmd.Stdout = buf
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return nil, err
				}
				var vr VaultResponse
				if err := json.Unmarshal(buf.Bytes(), &vr); err != nil {
					return nil, err
				}
				return vr.Data.Data, nil
			},
		},
	}

	r.rioconfig = &rio.Config{}

	flag.BoolVar(&r.dry, "d", false, "Dry run; Don't make any changes")
	flag.BoolVar(&r.diff, "D", false, "Show full diff of file content changes")
	flag.BoolVar(&r.verbose, "v", false, "Be more verbose")

	flag.StringVar(&r.host, "host", "", "Execute on host via SSH")
	flag.StringVar(&r.user, "user", os.Getenv("USER"), "User to SSH as")

	flag.Parse()

	r.rioconfig.Verbose = r.verbose

	title := "███ "

	if r.dry {
		title += "Dry running"
	} else {
		title += color(Green) + "Applying" + reset()
	}
	title += " " + brightcolor(Yellow) + describe + reset()

	if r.host == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		title += " via SSH on " + hostname
	} else {
		title += " locally on " + r.host
	}
	//title += " ..."
	fmt.Println(title)

	if r.host != "" {
		socket := os.Getenv("SSH_AUTH_SOCK")
		conn, err := net.Dial("unix", socket)
		if err != nil {
			return fmt.Errorf("Failed to open SSH_AUTH_SOCK: %w", err)
		}
		agentClient := agent.NewClient(conn)
		sshconfig := &ssh.ClientConfig{
			User: r.user,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeysCallback(agentClient.Signers),
			},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				// TODO
				return nil
			},
			BannerCallback: ssh.BannerDisplayStderr(),
		}

		r.rioconfig.Pool = sshpool.New(sshconfig, &sshpool.PoolConfig{Debug: r.verbose})
		r.rioconfig.Host = r.host
	}

	//	out := &outputter{}
	//	r.out = out

	//	total := len(items)
	//	finished := 0
	//	defer func() {
	//		fmt.Printf("%d/%d things up to date\n", finished, total)
	//	}()

	//	bar := progress.NewBar(total, "")
	//	defer bar.Done()
	//	out.bar = bar

	return r.run()
}
