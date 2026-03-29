package flow

import (
	"context"

	"github.com/oesand/octo/mediator"
)

// Event represents an external or internal occurrence that can be
// handled by flow steps. Events must expose the originating UID and
// the flow they belong to.
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

// TriggerEvent constructs a flow trigger event for the given UID and
// flow name. It can be published to cause the flow manager to resume
// processing for that UID.
func TriggerEvent(uid, flow string) Event {
	return &triggerEvent{uid, flow}
}

// Trigger publishes a trigger event to the provided mediator manager
// causing the flow for the specified UID to be processed.
func Trigger(manager *mediator.Manager, ctx context.Context, uid, flow string) error {
	return mediator.Publish(manager, ctx, TriggerEvent(uid, flow))
}
