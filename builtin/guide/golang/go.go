// Package files contains a component for validating local files.
package golang

import (
	"github.com/hashicorp/waypoint/sdk"
)

//go:generate protoc -I ../../../.. --go_opt=plugins=grpc --go_out=../../../.. waypoint/builtin/guide/golang/plugin.proto

// Options are the SDK options to use for instantiation for
// the Files plugin.
var Options = []sdk.Option{
	sdk.WithComponents(&Builder{}),
}
