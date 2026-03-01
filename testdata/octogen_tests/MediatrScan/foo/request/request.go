package request

import "context"

type ExampleRequest struct {
	Name string `json:"name"`
}

type ExampleResponse struct {
	Code int `json:"code"`
}

type ReqHandler struct {
	Stct *Struct
	Oth  *Other
}

func (*ReqHandler) Request(ctx context.Context, r *ExampleRequest) (*ExampleResponse, error) {
	return &ExampleResponse{Code: 1}, nil
}
