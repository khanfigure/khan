package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"net"
	"os"

	"github.com/desops/khan/host"
	"github.com/desops/khan/host/dry"
	"github.com/desops/khan/host/local"
	"github.com/desops/khan/host/remote"

	"github.com/desops/sshpool"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "✓ Pass")
}

func run() error {
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

	pool := sshpool.New(sshconfig, &sshpool.PoolConfig{Debug: true})
	defer pool.Close()

	var hosts []host.Host

	sugo := remote.New(pool, "root@sugo:1722")

	hosts = append(hosts, local.New())
	//	hosts = append(hosts, remote.New(pool, "joel.jensen@stan"))
	//	hosts = append(hosts, sugo)
	hosts = append(hosts, dry.New(0, 0, sugo))

	for _, host := range hosts {
		fmt.Println("doing host", host)
		if err := wr(host, "/tmp/file0"); err != nil {
			return err
		}
		if err := wr(host, "/tmp/file1"); err != nil {
			return err
		}
		if err := host.Remove("/tmp/file0"); err != nil {
			return err
		}
		if err := rd(host, "/tmp/file0"); err == nil {
			return fmt.Errorf("this file is supposed to be gone")
		}
		if err := rd(host, "/tmp/file2"); err != nil {
			return err
		}
		if err := host.Remove("/tmp/file2"); err != nil {
			return err
		}
		if err := rd(host, "/tmp/file2"); err == nil {
			return fmt.Errorf("this file is supposed to be gone")
		}
	}
	return nil
}

func rd(host host.Host, fpath string) error {
	content := "hi " + fpath + "\n"

	buf, err := host.ReadFile(fpath)
	//fmt.Printf("host.ReadFile(%#v) %v, %v\n", fpath, buf, err)
	if err != nil {
		return err
	}

	if string(buf) != content {
		return fmt.Errorf("Contents are different: %#v → %#v", content, string(buf))
	}

	return nil
}

func wr(host host.Host, fpath string) error {
	content := "hi " + fpath + "\n"

	fh, err := host.Create(fpath)
	if err != nil {
		return err
	}
	defer fh.Close()

	if _, err := fmt.Fprint(fh, content); err != nil {
		return err
	}
	if err := fh.Close(); err != nil {
		return err
	}

	// now read back
	buf, err := host.ReadFile(fpath)
	if err != nil {
		return err
	}

	if string(buf) != content {
		return fmt.Errorf("Contents are different: %#v → %#v", content, string(buf))
	}

	return nil
}
