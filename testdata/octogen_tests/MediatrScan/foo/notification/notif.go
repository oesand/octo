package notification

import (
	"context"

	"github.com/oesand/octo/testdata/octogen_tests/MediatrScan/foo/request"
)

type ExampleEvent struct {
	Name string `json:"name"`
}

type NotificationHandler struct {
	Stct *request.Struct
	Oth  *request.Other
}

func (*NotificationHandler) Notification(ctx context.Context, ev *ExampleEvent) error {
	return nil
}
