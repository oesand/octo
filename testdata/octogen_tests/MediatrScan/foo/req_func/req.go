package req_func

import (
	"context"
	"github.com/oesand/octo/testdata/octogen_tests/MediatrScan/foo/req"
)

func NewReqHandler() *ReqHandler {
	return &ReqHandler{}
}

type ExampleReq struct {
	Name string `json:"name"`
}

type ExampleResp struct {
	Code int `json:"code"`
}

type ReqHandler struct {
	Stct *req.Struct
	Oth  *req.Other
}

func (*ReqHandler) Request(ctx context.Context, r *ExampleReq) (*ExampleResp, error) {
	return &ExampleResp{Code: 1}, nil
}
