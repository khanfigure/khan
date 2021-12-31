package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	gobuild "go/build"

	"github.com/go-bindata/go-bindata/v3"
)

type buildrun struct {
	staticfiles []string
	wd          string
	cwd         string
}

func build() error {

	describe := "unknown"

	if _, err := os.Stat(".git"); err == nil {
		cmd := exec.Command("git", "describe", "--all", "--long", "--dirty")
		buf, err := cmd.Output()
		if err == nil {
			describe = string(buf)
		}
		// ignore err on purpose. maybe you aren't in a git repo.
	}

	if _, err := os.Stat("go.mod"); err != nil {
		return fmt.Errorf("Uninitialized khan configuration, or not executed from project root. Initialize with: khan init")
	}

	buf, err := exec.Command("go", "list", "-m").Output()
	if err != nil {
		return fmt.Errorf("go list -m failed: %w", err)
	}

	outfile := strings.TrimSpace(string(buf))

	title := outfile

	// Enter a private space in /tmp so that we don't clutter the cwd with
	// generated intermediate crap. This space will persist: it is based on
	// the working directory absolute path. This way the go compiler can
	// optimize multiple runs.

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// safe, clean, slow way
	//	wd, err := ioutil.TempDir("", "khan")
	//	if err != nil {
	//		return err
	//	}

	// more fun way
	wd, err := stabletmpdir(cwd)
	if err != nil {
		return err
	}

	br := &buildrun{
		wd:  wd,
		cwd: cwd,
	}

	fmt.Println("Building", wd, "→", outfile)

	copybacklist := []string{"go.sum", "go.mod"}

	if err := sync(wd, cwd); err != nil {
		return err
	}

	matches, err := filepath.Glob("*.yaml")
	if err != nil {
		return err
	}
	matches2, err := filepath.Glob("*.yml")
	if err != nil {
		return err
	}
	matches = append(matches, matches2...)
	sort.Strings(matches)

	assetfs := false

	for _, match := range matches {
		base := filepath.Base(match)
		goname := base + ".go"
		if err := br.yaml2go(wd, match, wd+"/"+goname, &assetfs); err != nil {
			return err
		}
	}

	bc := bindata.NewConfig()
	bc.Output = br.wd + "/go_bindata_static_files.go"
	bc.Package = "main"
	staticfiledups := map[string]bool{}
	if len(br.staticfiles) > 0 {
		for _, file := range br.staticfiles {
			file = filepath.Clean(file)
			if staticfiledups[file] {
				continue
			}
			staticfiledups[file] = true
			bc.Input = append(bc.Input, bindata.InputConfig{
				Path:      file,
				Recursive: false,
			})
		}
		fmt.Println("Bundling", len(br.staticfiles), "static files ...")
	}

	if err := bindata.Translate(bc); err != nil {
		return err
	}

	if _, err := os.Stat(wd + "/go.mod"); err != nil {
		if err := ioutil.WriteFile(wd+"/go.mod", []byte(`module myconfig
`), 0644); err != nil {
			return err
		}
	}

	if _, err := os.Stat(wd + "/main.go"); err != nil {
		if err := ioutil.WriteFile(wd+"/main.go", []byte(fmt.Sprintf(`package main
import (
	"fmt"
	"os"
	"io"
	"bytes"

	%s %#v
)

type fakecloser struct {
	io.Reader
}
func (fc fakecloser) Close() error {
	return nil
}
func assetfn(path string) (io.ReadCloser, error) {
	buf, err := Asset(path)
	if err != nil {
		return nil, err
	}
	return fakecloser{bytes.NewReader(buf)}, nil
}

func main() {
	%s.SetTitle(%#v)
	%s.SetSourcePrefix(%#v)
	%s.SetDescribe(%#v)
	%s.SetAssetLoader(assetfn)

	if err := %s.Apply(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
`, khanpkgalias, khanpkgname, khanpkgalias, title, khanpkgalias, wd, khanpkgalias, strings.TrimSpace(describe), khanpkgalias, khanpkgalias)), 0644); err != nil {
			return err
		}
	}

	// before we compile: We need to recursively look for import paths, and do any additional yaml translation for those packages

	buildcontext := gobuild.Default
	buildcontext.Dir = wd
	fmt.Printf("ImportDir(%#v)\n", wd)
	pkg, err := buildcontext.ImportDir(wd, 0)
	if err != nil {
		return err
	}
	fmt.Printf("package %#v at %#v\n", pkg.Name, pkg.Dir)
	for _, gof := range pkg.GoFiles {
		fmt.Println(gof)
	}
	for _, imp := range pkg.Imports {
		fmt.Println("->", imp)

		ipkg, err := buildcontext.Import(imp, wd, 0)
		if err != nil {
			return err
		}

		fmt.Println("        ", ipkg.Dir, ipkg.ImportPath)
	}


	//fmt.Println("Compiling ...")

	cmd := exec.Command("go", "build", "-o", cwd+"/"+outfile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = wd
	if err := cmd.Run(); err != nil {
		return err
	}

	// Copy back out files that go often changes, but only if you had them in your original
	// working directory. If you didn't have the file originally, you probably don't care.
	for _, p := range copybacklist {
		if _, err := os.Stat(cwd + "/" + p); err == nil {
			err = compare(cwd+"/"+p, wd+"/"+p)
			if err == nil {
				continue
			}
			if err == errNotSame {
				if err := cp(cwd+"/"+p, wd+"/"+p); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}

func copyglobs(dest string, globs ...string) error {
	for _, g := range globs {
		matches, err := filepath.Glob(g)
		if err != nil {
			return err
		}
		for _, match := range matches {
			base := filepath.Base(match)
			//dir := filepath.Dir(match)
			to := dest + "/" + base

			destfh, err := os.Create(to)
			if err != nil {
				return err
			}
			defer destfh.Close()
			srcfh, err := os.Open(match)
			if err != nil {
				return err
			}
			defer srcfh.Close()
			if _, err := io.Copy(destfh, srcfh); err != nil {
				return err
			}
			if err := destfh.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

func stabletmpdir(srcpath string) (string, error) {
	p, err := filepath.EvalSymlinks(srcpath)
	if err != nil {
		return "", err
	}

	h := fmt.Sprintf("%x", md5.Sum([]byte(p)))
	r := "/tmp/khan_" + h[:8]

	info, err := os.Stat(r)
	if err == nil && info.IsDir() {
		return r, nil
	}

	//fmt.Println("mkdir", r)
	if err := os.Mkdir(r, 0700); err != nil {
		return "", err
	}
	return r, nil
}
