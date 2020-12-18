package duck

import (
	"io"
	"path/filepath"
	//"time"

	"github.com/flosch/pongo2/v4"
)

type bindataloader struct {
	run *run
}

func (bdl *bindataloader) Abs(base, name string) string {
	return filepath.Join(base, name)
}
func (bdl *bindataloader) Get(path string) (io.Reader, error) {
	return bdl.run.assetfn(path)
}

func executeTemplate(r *run, tfile string) (string, error) {

	//time.Sleep(time.Second)

	ts := pongo2.NewSet("go-bindata", &bindataloader{r})

	tpl, err := ts.FromFile(tfile)
	if err != nil {
		return "", err
	}
	buf, err := tpl.ExecuteBytes(pongo2.Context{"tier": "prod"})
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
