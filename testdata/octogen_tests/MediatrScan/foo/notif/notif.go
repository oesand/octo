package notif

import (
	"context"
	"github.com/oesand/octo/testdata/octogen_tests/MediatrScan/foo/req"
)

type ExampleEvent struct {
	Name string `json:"name"`
}

type NotifHandler struct {
	Stct *req.Struct
	Oth  *req.Other
}

func (*NotifHandler) Notification(ctx context.Context, ev *ExampleEvent) {
}
