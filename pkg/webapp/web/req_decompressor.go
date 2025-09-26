package web

// Decompressor 请求解压管理者
type ReqDecompressor interface {
	Decompression(tName string) (ReqDecompression, bool)
}
