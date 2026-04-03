package reqdecp

import "github.com/xiaohangshu-dev/go-workit/pkg/webapp/web"

type ReqDecompressor struct {
	reqDecompression map[string]web.ReqDecompression
}

func NewReqDecompressor(reqDecompressor map[string]web.ReqDecompression) *ReqDecompressor {

	return &ReqDecompressor{
		reqDecompression: reqDecompressor,
	}

}

func (rd *ReqDecompressor) Decompression(tName string) (web.ReqDecompression, bool) {

	val, exits := rd.reqDecompression[tName]

	if exits {
		return val, true
	}

	return nil, false

}
