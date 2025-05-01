package framework

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// Helper functions for proto timestamp conversion
func timestampProto(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func timeFromProto(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}
