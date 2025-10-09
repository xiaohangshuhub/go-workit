package br

import (
	"io"

	"github.com/andybalholm/brotli"
)

type Brotli struct {
}

func New() *Brotli {
	return &Brotli{}
}

func (b *Brotli) Type() string {
	return "br"
}

func (b *Brotli) Decompression(r io.Reader) (io.ReadCloser, error) {

	return io.NopCloser(brotli.NewReader(r)), nil
}
