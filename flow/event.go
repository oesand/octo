package flow

import (
	"context"

	"github.com/oesand/octo/mediator"
)

type Event interface {
	Uid() string
	Flow() string
}

type triggerEvent struct {
	uid  string
	flow string
}

func (ev *triggerEvent) Uid() string {
	return ev.uid
}

func (ev *triggerEvent) Flow() string {
	return ev.flow
}

func TriggerEvent(uid, flow string) Event {
	return &triggerEvent{uid, flow}
}

func Trigger(manager *mediator.Manager, ctx context.Context, uid, flow string) error {
	return mediator.Publish(manager, ctx, TriggerEvent(uid, flow))
}
