package gzip

import (
	"compress/gzip"
	"io"
)

type Gzip struct {
}

func New() *Gzip {
	return &Gzip{}
}

func (b *Gzip) Type() string {
	return "gzip"
}

func (b *Gzip) Decompression(r io.Reader) (io.ReadCloser, error) {

	return gzip.NewReader(r)
}
