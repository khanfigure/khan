package khan

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/flosch/pongo2/v4"
)

type VaultResponse struct {
	Data VaultResponseData
}
type VaultResponseData struct {
	Data map[string]string
}

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
	ts := pongo2.NewSet("go-bindata", &bindataloader{r})

	tpl, err := ts.FromFile(tfile)
	if err != nil {
		return "", err
	}
	buf, err := tpl.ExecuteBytes(pongo2.Context{
		"khan": map[string]interface{}{
			"secret": func(path string) (map[string]string, error) {
				buf := &bytes.Buffer{}
				cmd := r.rioconfig.Command(context.Background(), "vault", "kv", "get", "-format", "json", "secret/"+path)
				cmd.Stdout = buf
				cmd.Stderr = os.Stderr
				cmd.Env = [][2]string{
					{"VAULT_ADDR", "http://localhost:8200"},
				}
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
	})
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
