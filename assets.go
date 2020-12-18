package duck

import (
	"bytes"
	"io"
	"os"
)

// this will either be a no-op loader or an assetfs loader,
// depending on if files are embedded with go-bindata or not.
var mainassetfn func(string) (io.Reader, error)

func SetAssetLoader(fn func(string) (io.Reader, error)) {
	mainassetfn = fn
}

func dummyassetfn(_ string) (io.Reader, error) {
	_ = bytes.NewReader
	return nil, os.ErrNotExist
}
