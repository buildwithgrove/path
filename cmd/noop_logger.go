package cmd

import (
	"context"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
)

// noOpLogger is a Logger implementation that performs no operations.
// All methods return no-op implementations or the receiver itself.
type noOpLogger struct{}

// NewNoOpLogger creates a new no-operation logger.
func NewNoOpLogger() polylog.Logger {
	return &noOpLogger{}
}

// Debug returns a no-op Event.
func (n *noOpLogger) Debug() polylog.Event {
	return &noOpEvent{}
}

// ProbabilisticDebugInfo returns a no-op Event, ignoring the probability.
func (n *noOpLogger) ProbabilisticDebugInfo(float64) polylog.Event {
	return &noOpEvent{}
}

// Info returns a no-op Event.
func (n *noOpLogger) Info() polylog.Event {
	return &noOpEvent{}
}

// Warn returns a no-op Event.
func (n *noOpLogger) Warn() polylog.Event {
	return &noOpEvent{}
}

// Error returns a no-op Event.
func (n *noOpLogger) Error() polylog.Event {
	return &noOpEvent{}
}

// With returns the same no-op logger, ignoring the key-value pairs.
func (n *noOpLogger) With(keyVals ...any) polylog.Logger {
	return n
}

// WithContext returns the context unchanged.
func (n *noOpLogger) WithContext(ctx context.Context) context.Context {
	return ctx
}

// WithLevel returns a no-op Event, ignoring the level.
func (n *noOpLogger) WithLevel(level polylog.Level) polylog.Event {
	return &noOpEvent{}
}

// Write implements io.Writer by doing nothing and returning the length of p.
func (n *noOpLogger) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// noOpEvent is an Event implementation that performs no operations.
// All methods return the receiver itself or appropriate default values.
type noOpEvent struct{}

// Str returns the receiver, ignoring the key-value pair.
func (e *noOpEvent) Str(key, value string) polylog.Event {
	return e
}

// Bool returns the receiver, ignoring the key-value pair.
func (e *noOpEvent) Bool(key string, value bool) polylog.Event {
	return e
}

// Int returns the receiver, ignoring the key-value pair.
func (e *noOpEvent) Int(key string, value int) polylog.Event {
	return e
}

// Int8 returns the receiver, ignoring the key-value pair.
func (e *noOpEvent) Int8(key string, value int8) polylog.Event {
	return e
}

// Int16 returns the receiver, ignoring the key-value pair.
func (e *noOpEvent) Int16(key string, value int16) polylog.Event {
	return e
}

// Int32 returns the receiver, ignoring the key-value pair.
func (e *noOpEvent) Int32(key string, value int32) polylog.Event {
	return e
}

// Int64 returns the receiver, ignoring the key-value pair.
func (e *noOpEvent) Int64(key string, value int64) polylog.Event {
	return e
}

// Uint returns the receiver, ignoring the key-value pair.
func (e *noOpEvent) Uint(key string, value uint) polylog.Event {
	return e
}

// Uint8 returns the receiver, ignoring the key-value pair.
func (e *noOpEvent) Uint8(key string, value uint8) polylog.Event {
	return e
}

// Uint16 returns the receiver, ignoring the key-value pair.
func (e *noOpEvent) Uint16(key string, value uint16) polylog.Event {
	return e
}

// Uint32 returns the receiver, ignoring the key-value pair.
func (e *noOpEvent) Uint32(key string, value uint32) polylog.Event {
	return e
}

// Uint64 returns the receiver, ignoring the key-value pair.
func (e *noOpEvent) Uint64(key string, value uint64) polylog.Event {
	return e
}

// Float32 returns the receiver, ignoring the key-value pair.
func (e *noOpEvent) Float32(key string, value float32) polylog.Event {
	return e
}

// Float64 returns the receiver, ignoring the key-value pair.
func (e *noOpEvent) Float64(key string, value float64) polylog.Event {
	return e
}

// Err returns the receiver, ignoring the error.
func (e *noOpEvent) Err(err error) polylog.Event {
	return e
}

// Timestamp returns the receiver, ignoring the timestamp.
func (e *noOpEvent) Timestamp() polylog.Event {
	return e
}

// Time returns the receiver, ignoring the key-value pair.
func (e *noOpEvent) Time(key string, value time.Time) polylog.Event {
	return e
}

// Dur returns the receiver, ignoring the key-value pair.
func (e *noOpEvent) Dur(key string, value time.Duration) polylog.Event {
	return e
}

// Fields returns the receiver, ignoring the fields.
func (e *noOpEvent) Fields(fields any) polylog.Event {
	return e
}

// Func returns the receiver, ignoring the function.
func (e *noOpEvent) Func(func(polylog.Event)) polylog.Event {
	return e
}

// Enabled always returns false since this is a no-op implementation.
func (e *noOpEvent) Enabled() bool {
	return false
}

// Discard returns the receiver.
func (e *noOpEvent) Discard() polylog.Event {
	return e
}

// Msg does nothing.
func (e *noOpEvent) Msg(message string) {
	// no-op
}

// Msgf does nothing.
func (e *noOpEvent) Msgf(format string, keyVals ...interface{}) {
	// no-op
}

// Send does nothing.
func (e *noOpEvent) Send() {
	// no-op
}