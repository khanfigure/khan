package host

import (
	"io"
)

type Host interface {
	String() string

	Create(string) (io.WriteCloser, error)
	Remove(string) error

	Open(string) (io.ReadCloser, error)
	ReadFile(string) ([]byte, error)
}
