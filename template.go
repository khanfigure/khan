package khan

import (
	"os"
	"io"
	"path/filepath"
	"bytes"
	"context"
	"encoding/json"
)

type VaultResponse struct {
	Data VaultResponseData
}
type VaultResponseData struct {
	Data map[string]string
}

type bindataloader struct {
	run *Run
}

func (bdl *bindataloader) Abs(base, name string) string {
	return filepath.Join(base, name)
}
func (bdl *bindataloader) Get(path string) (io.Reader, error) {
	return bdl.run.assetfn(path)
}

func setContextHostTools(pcontext map[string]interface{}, host *Host) {
	kh := pcontext["khan"].(map[string]interface{})
	kh["secret"] = func(path string) (map[string]string, error) {
				buf := &bytes.Buffer{}
				cmd := host.Command(context.Background(), "vault", "kv", "get", "-format", "json", "secret/"+path)
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
			}
}

func executePackedTemplateFile(host *Host, tfile string) (string, error) {
	host.Run.pongomu.Lock()
	defer host.Run.pongomu.Unlock()

	v, ok := host.Run.pongocachefiles[tfile]

	if !ok {
		tpl, err := host.Run.pongopackedset.FromFile(tfile)
		if err != nil {
			return "", err
		}
		host.Run.pongocachefiles[tfile] = tpl
		v = tpl
	}

	setContextHostTools(host.Run.pongopackedcontext, host)

	buf, err := v.ExecuteBytes(host.Run.pongopackedcontext)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func executePackedTemplateString(host *Host, s string) (string, error) {
	host.Run.pongomu.Lock()
	defer host.Run.pongomu.Unlock()

	v, ok := host.Run.pongocachestrings[s]
	if !ok {
		tpl, err := host.Run.pongopackedset.FromString(s)
		if err != nil {
			return "", err
		}
		host.Run.pongocachestrings[s] = tpl
		v = tpl
	}

	setContextHostTools(host.Run.pongopackedcontext, host)

	buf, err := v.ExecuteBytes(host.Run.pongopackedcontext)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
