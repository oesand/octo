package ocp_test

import (
	"testing"

	"github.com/oesand/octo"
	"github.com/oesand/octo/ocp"
)

var _ = ocp.Collect("main",
	ocp.Inject2(octo.Resolve[*First], octo.Resolve[*Second], NewStruct),
	ocp.InjectR2(NewStruct),
)

func NewStruct(first *First, second *Second) *Struct {
	return &Struct{}
}

type First struct{}
type Second struct{}
type Struct struct{}

func TestCollectGroup(t *testing.T) {
	container := octo.New()

	ocp.InjectGroup(container, "main")

}
