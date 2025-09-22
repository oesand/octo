package mediatr

import (
	"context"
	"errors"
	"github.com/oesand/octo"
	"strings"
	"testing"
)

// --- Example request/response types and handlers ---

type EchoRequest struct {
	Message string
}

type EchoResponse struct {
	Reply string
}

type EchoHandler struct{}

func (h *EchoHandler) Request(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	if req.Message == "" {
		return EchoResponse{}, errors.New("empty message")
	}
	return EchoResponse{Reply: "Echo: " + req.Message}, nil
}

type SumRequest struct {
	A, B int
}

type SumResponse struct {
	Result int
}

type SumHandler struct{}

func (h *SumHandler) Request(ctx context.Context, req SumRequest) (SumResponse, error) {
	return SumResponse{Result: req.A + req.B}, nil
}

type FailRequest struct{}

type FailResponse struct{}

type FailHandler struct{}

func (h *FailHandler) Request(ctx context.Context, req FailRequest) (FailResponse, error) {
	return FailResponse{}, errors.New("always fails")
}

// --- Tests ---

func TestInjectValueAndSend_Success(t *testing.T) {
	c := octo.New()

	octo.InjectValue(c, &EchoHandler{})

	resp, err := Send[EchoRequest, EchoResponse](c, context.Background(), EchoRequest{Message: "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Reply != "Echo: hello" {
		t.Fatalf("expected 'Echo: hello', got '%s'", resp.Reply)
	}
}

func TestInjectAndSend_Success(t *testing.T) {
	c := octo.New()

	octo.Inject(c, func(container *octo.Container) *EchoHandler {
		return &EchoHandler{}
	})

	resp, err := Send[EchoRequest, EchoResponse](c, context.Background(), EchoRequest{Message: "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Reply != "Echo: hello" {
		t.Fatalf("expected 'Echo: hello', got '%s'", resp.Reply)
	}
}

func TestInjectAndSend_Error(t *testing.T) {
	c := octo.New()

	octo.InjectValue(c, &EchoHandler{})

	resp, err := Send[EchoRequest, EchoResponse](c, context.Background(), EchoRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if resp.Reply != "" {
		t.Fatalf("expected empty reply, got '%s'", resp.Reply)
	}
}

func TestSend_PanicsWhenHandlerNotRegistered(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when no handler is registered")
		}
		if !strings.HasPrefix(r.(string), "octo: fail to resolve type mediatr.RequestHandler") {
			t.Fatalf("unexpected panic message, got '%s'", r)
		}
	}()
	c := octo.New()
	// Nothing registered
	_, _ = Send[EchoRequest, EchoResponse](c, context.Background(), EchoRequest{Message: "fail"})
}

func TestInjectManyDifferentHandlers(t *testing.T) {
	c := octo.New()

	octo.InjectValue(c, &EchoHandler{})
	octo.InjectValue(c, &SumHandler{})
	octo.InjectValue(c, &FailHandler{})

	// Test Echo
	echoResp, err := Send[EchoRequest, EchoResponse](c, context.Background(), EchoRequest{Message: "test"})
	if err != nil {
		t.Fatalf("unexpected error for EchoRequest: %v", err)
	}
	if echoResp.Reply != "Echo: test" {
		t.Fatalf("expected 'Echo: test', got '%s'", echoResp.Reply)
	}

	// Test Sum
	sumResp, err := Send[SumRequest, SumResponse](c, context.Background(), SumRequest{A: 2, B: 3})
	if err != nil {
		t.Fatalf("unexpected error for SumRequest: %v", err)
	}
	if sumResp.Result != 5 {
		t.Fatalf("expected 5, got %d", sumResp.Result)
	}

	// Test Fail
	_, err = Send[FailRequest, FailResponse](c, context.Background(), FailRequest{})
	if err == nil {
		t.Fatal("expected error for FailRequest, got nil")
	}
}
