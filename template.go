package khan

import (
	"io"
	"path/filepath"
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

func executePackedTemplateFile(r *run, tfile string) (string, error) {
	v, ok := r.pongocachefiles[tfile]

	if !ok {
		tpl, err := r.pongopackedset.FromFile(tfile)
		if err != nil {
			return "", err
		}
		r.pongocachefiles[tfile] = tpl
		v = tpl
	}

	buf, err := v.ExecuteBytes(r.pongopackedcontext)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func executePackedTemplateString(r *run, s string) (string, error) {
	v, ok := r.pongocachestrings[s]
	if !ok {
		tpl, err := r.pongopackedset.FromString(s)
		if err != nil {
			return "", err
		}
		r.pongocachestrings[s] = tpl
		v = tpl
	}

	buf, err := v.ExecuteBytes(r.pongopackedcontext)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
