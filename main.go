package khan

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/desops/khan/rio"

	"github.com/desops/sshpool"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var (
	describe     string
	sourceprefix string
)

func SetSourcePrefix(s string) {
	sourceprefix = s
}
func SetDescribe(s string) {
	describe = s
}

func Apply() error {
	r := &run{
		assetfn: mainassetfn,
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
		title += "Applying"
	}
	title += " " + brightcolor(Yellow) + describe + reset() + " on "

	if r.host == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		title += " " + hostname
	} else {
		title += " " + r.host
	}
	title += "..."
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

	out := &outputter{}
	r.out = out

	total := len(items)
	finished := 0

	defer func() {
		fmt.Printf("%d/%d things up to date\n", finished, total)
	}()

	for _, item := range items {
		out.StartItem(r, item)

		status, err := item.apply(r)

		out.FinishItem(r, item, status, err)

		if err != nil {
			// wrap the error with its source
			md := meta[item.getID()]
			err = fmt.Errorf("%s %w", strings.TrimPrefix(md.source, sourceprefix+"/"), err)
			return err
		}

		if !r.dry || status == itemUnchanged {
			finished++
		}
	}

	return nil
}
