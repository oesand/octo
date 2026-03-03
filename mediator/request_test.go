package mediator_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/oesand/octo"
	"github.com/oesand/octo/mediator"
)

type TestRequest struct {
	mediator.Request[TestResponse]
	Value int
}

type TestResponse struct {
	Result int
}

type TestRequestHandler struct {
	Called atomic.Bool
}

func (h *TestRequestHandler) Request(_ context.Context, req TestRequest) (TestResponse, error) {
	h.Called.Store(true)
	return TestResponse{Result: req.Value * 2}, nil
}

func TestSend_RequestHandler(t *testing.T) {
	container := octo.New()
	handler1 := &TestRequestHandler{}
	octo.InjectValue(container, handler1)

	handler2 := &TestRequestHandler{}
	octo.InjectValue(container, handler2)

	manager := mediator.Inject(container)

	resp, err := mediator.Send(manager, context.Background(), TestRequest{Value: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Result != 6 {
		t.Fatalf("expected 6, got %d", resp.Result)
	}

	if !handler1.Called.Load() {
		t.Fatal("expected handler1 called")
	}

	if handler2.Called.Load() {
		t.Fatal("expected handler2 not called")
	}
}
