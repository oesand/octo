package foo

import (
	"context"
	"github.com/oesand/octo/mediatr"
)

type Event struct{}

type Req struct{}
type Resp struct{}

// Basic structs

type Iface interface {
	mediatr.NotificationHandler[*Event]
}

type NotFuncStruct struct{}

func (*NotFuncStruct) Notification(ctx context.Context, ev *Event) error {}

type InvalidReturnsStruct struct{}

func (*InvalidReturnsStruct) Notification(ctx context.Context, ev *Event) error {}

type GenericFuncStruct struct{}

func (*GenericFuncStruct) Request(ctx context.Context, request *Req) (*Resp, error) {
	return &Resp{}, nil
}

type FuncReturnNotStruct struct{}

func (*FuncReturnNotStruct) Request(ctx context.Context, request *Req) (*Resp, error) {
	return &Resp{}, nil
}

type InvalidFuncParamStruct struct{}

func (*InvalidFuncParamStruct) Request(ctx context.Context, request *Req) (*Resp, error) {
	return &Resp{}, nil
}

// Warning types

type GenericStruct[T any] struct{}

func (*GenericStruct[T]) Notification(ctx context.Context, ev *Event) error {}

type InvalidFieldStruct struct {
	Fld0 string
	Fld1 int
}

func (*InvalidFieldStruct) Notification(ctx context.Context, ev *Event) error {}

var NewNotFuncStruct int

func NewInvalidReturnsStruct() (*InvalidReturnsStruct, error) {
	return &InvalidReturnsStruct{}, nil
}

func NewGenericFuncStruct[T any]() *GenericFuncStruct {
	return &GenericFuncStruct{}
}

func NewFuncReturnNotStruct() string {
	return ""
}

func NewInvalidFuncParamStruct(p0 string, p1 int) *InvalidFuncParamStruct {
	return &InvalidFuncParamStruct{}
}
