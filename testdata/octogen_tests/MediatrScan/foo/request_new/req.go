package request_new

import (
	"context"
	"github.com/oesand/octo/testdata/octogen_tests/MediatrScan/foo/request"
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
	Stct *request.Struct
	Oth  *request.Other
}

func (*ReqHandler) Request(ctx context.Context, r *ExampleReq) (*ExampleResp, error) {
	return &ExampleResp{Code: 1}, nil
}
