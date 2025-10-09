package web

import "io"

// Decompression 请求解压接口
type ReqDecompression interface {
	Type() string
	Decompression(r io.Reader) (io.ReadCloser, error)
}
