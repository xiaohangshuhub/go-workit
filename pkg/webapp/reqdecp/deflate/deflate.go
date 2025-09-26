package deflate

import (
	"compress/zlib"
	"io"
)

type Deflate struct {
}

func New() *Deflate {
	return &Deflate{}
}

func (b *Deflate) Type() string {
	return "deflate"
}

func (b *Deflate) Decompression(r io.Reader) (io.ReadCloser, error) {

	return zlib.NewReader(r)
}
