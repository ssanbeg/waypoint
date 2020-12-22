package logviewer

import (
	"context"

	"github.com/golang/protobuf/ptypes"

	"github.com/hashicorp/waypoint-plugin-sdk/component"
	pb "github.com/hashicorp/waypoint/internal/server/gen"
)

// Viewer implements component.LogViewer over the server-side log stream endpoint.
//
// TODO(mitchellh): we should support some form of reconnection in the event of
// network errors.
type Viewer struct {
	// Stream is the log stream client to use.
	Stream pb.Waypoint_GetLogStreamClient
}

// NextLogBatch implements component.LogViewer
func (v *Viewer) NextLogBatch(ctx context.Context) ([]component.LogEvent, error) {
	// Get the next batch. Note that we specifically do NOT buffer here because
	// we want to provide the proper amount of backpressure and we expect our
	// downstream caller to be calling these as quickly as possible.
	batch, err := v.Stream.Recv()
	if err != nil {
		return nil, err
	}

	events := make([]component.LogEvent, len(batch.Lines))
	for i, entry := range batch.Lines {
		ts, _ := ptypes.Timestamp(entry.Timestamp)

		events[i] = component.LogEvent{
			Partition: batch.InstanceId,
			Timestamp: ts,
			Message:   entry.Line,
		}
	}

	return events, nil
}

var _ component.LogViewer = (*Viewer)(nil)
