package mediatr

import (
	"context"
	"github.com/oesand/octo"
	"testing"
)

type ExampleShortRequest struct {
	Name string
}

func (*ExampleShortRequest) Returns(*ExampleShortResponse) {}

type ExampleShortResponse struct {
	Code int
}

type ExampleShortRequestHandler struct {
	Last string
}

func (h *ExampleShortRequestHandler) Request(_ context.Context, r *ExampleShortRequest) (*ExampleShortResponse, error) {
	h.Last = r.Name
	return &ExampleShortResponse{
		Code: len(r.Name),
	}, nil
}

func TestSendShort(t *testing.T) {
	container := octo.New()
	ctx := context.Background()

	handler := &ExampleShortRequestHandler{}
	octo.InjectValue(container, handler)

	name := "ExampleShortRequest: test text"
	req := &ExampleShortRequest{Name: name}

	resp, err := SendShort(container, ctx, req)
	if err != nil {
		t.Error(err)
	}

	if resp.Code != len(name) {
		t.Errorf("got invalid code %d, want %d", resp.Code, len(name))
	}

	if handler.Last != name {
		t.Errorf("got invalid handler last %s, want %s", handler.Last, name)
	}
}
