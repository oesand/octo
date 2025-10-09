package req

import "context"

type ExampleReq struct {
	Name string `json:"name"`
}

type ExampleResp struct {
	Code int `json:"code"`
}

type ReqHandler struct {
	Stct *Struct
	Oth  *Other
}

func (*ReqHandler) Request(ctx context.Context, r *ExampleReq) (*ExampleResp, error) {
	return &ExampleResp{Code: 1}, nil
}
