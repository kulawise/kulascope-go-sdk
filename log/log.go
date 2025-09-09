package log

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type contextKey string

const (
	loggerKey  contextKey = "request_logger"
	traceIDKey contextKey = "trace_id"
)

// SubLogRequest represents an individual log entry stored per request
type SubLogRequest struct {
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Error     string         `json:"error,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// EventLogger is the per-request logger
type EventLogger struct {
	ctx     context.Context
	zlog    *zerolog.Logger
	mu      sync.Mutex
	logs    []SubLogRequest
	traceID uuid.UUID
}

// NewEventLogger creates a new EventLogger for a request
func NewEventLogger(ctx context.Context, base *zerolog.Logger, traceID uuid.UUID) *EventLogger {
	l := &EventLogger{
		ctx:     ctx,
		zlog:    base,
		logs:    []SubLogRequest{},
		traceID: traceID,
	}
	return l
}

// WithLogger stores the EventLogger in context
func WithLogger(ctx context.Context, l *EventLogger) context.Context {
	ctx = context.WithValue(ctx, loggerKey, l)
	ctx = WithTraceID(ctx, l.traceID)
	return ctx
}

// FromContext retrieves the EventLogger from context
func FromContext(ctx context.Context) *EventLogger {
	if l, ok := ctx.Value(loggerKey).(*EventLogger); ok {
		return l
	}
	// fallback: no logger set
	nop := zerolog.Nop()
	return NewEventLogger(ctx, &nop, uuid.Nil)
}

// Logs returns all captured logs for the request
func (l *EventLogger) Logs() []SubLogRequest {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]SubLogRequest(nil), l.logs...)
}

// appendLog adds a sublog entry to the buffer
func (l *EventLogger) appendLog(level, msg string, metadata map[string]any, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry := SubLogRequest{
		Level:     level,
		Message:   msg,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}
	if err != nil {
		entry.Error = err.Error()
	}
	l.logs = append(l.logs, entry)
}

// ---------------------- EventWrapper ----------------------

type EventWrapper struct {
	logger   *EventLogger
	event    *zerolog.Event
	level    string
	metadata map[string]any
	err      error
}

// newEventWrapper initializes a wrapped zerolog.Event
func (l *EventLogger) newEventWrapper(level string) *EventWrapper {
	return &EventWrapper{
		logger:   l,
		event:    l.zlog.WithLevel(parseLevel(level)).Str("trace_id", l.traceID.String()),
		level:    level,
		metadata: map[string]any{},
	}
}

// Str adds a string field
func (e *EventWrapper) Str(key, val string) *EventWrapper {
	e.metadata[key] = val
	e.event.Str(key, val)
	return e
}

// Int adds an int field
func (e *EventWrapper) Int(key string, val int) *EventWrapper {
	e.metadata[key] = val
	e.event.Int(key, val)
	return e
}

// Err adds an error
func (e *EventWrapper) Err(err error) *EventWrapper {
	if err != nil {
		e.err = err
		e.event.Err(err)
		e.metadata["error"] = err.Error()
	}
	return e
}

// Msg writes the log and stores it in sublogs
func (e *EventWrapper) Msg(msg string) {
	e.logger.appendLog(e.level, msg, e.metadata, e.err)
	e.event.Msg(msg)
}

// Msgf writes a formatted log message and stores it in sublogs
func (e *EventWrapper) Msgf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	e.logger.appendLog(e.level, msg, e.metadata, e.err)
	e.event.Msg(msg)
}

func (e *EventWrapper) Interface(key string, val any) *EventWrapper {
	e.metadata[key] = val
	e.event.Interface(key, val) // call underlying zerolog
	return e
}

// Bool adds a boolean field to the event
func (e *EventWrapper) Bool(key string, val bool) *EventWrapper {
	e.metadata[key] = val
	e.event.Bool(key, val)
	return e
}

// Float64 adds a float64 field
func (e *EventWrapper) Float64(key string, val float64) *EventWrapper {
	e.metadata[key] = val
	e.event.Float64(key, val)
	return e
}

// Dur adds a time.Duration field
func (e *EventWrapper) Dur(key string, val time.Duration) *EventWrapper {
	e.metadata[key] = val
	e.event.Dur(key, val)
	return e
}

// Time adds a time.Time field
func (e *EventWrapper) Time(key string, val time.Time) *EventWrapper {
	e.metadata[key] = val
	e.event.Time(key, val)
	return e
}

func (e *EventWrapper) RawJSON(key string, val []byte) *EventWrapper {
	e.metadata[key] = string(val)
	e.event.RawJSON(key, val)
	return e
}

// An example of adding nested object support (like Dict in zerolog)
func (e *EventWrapper) Dict(key string, dict *zerolog.Event) *EventWrapper {
	e.event.Dict(key, dict)
	return e
}

// Any is an alias for Interface
func (e *EventWrapper) Any(key string, val any) *EventWrapper {
	return e.Interface(key, val)
}

// Strs adds a slice of strings field to the event
func (e *EventWrapper) Strs(key string, val []string) *EventWrapper {
	e.metadata[key] = val
	e.event.Strs(key, val)
	return e
}

// ---------------------- Convenience methods ----------------------

func (l *EventLogger) Info() *EventWrapper  { return l.newEventWrapper("info") }
func (l *EventLogger) Debug() *EventWrapper { return l.newEventWrapper("debug") }
func (l *EventLogger) Warn() *EventWrapper  { return l.newEventWrapper("warn") }
func (l *EventLogger) Error() *EventWrapper { return l.newEventWrapper("error") }
func (l *EventLogger) Trace() *EventWrapper { return l.newEventWrapper("trace") }
func (l *EventLogger) Fatal() *EventWrapper { return l.newEventWrapper("fatal") }
func (l *EventLogger) Panic() *EventWrapper { return l.newEventWrapper("panic") }

// ---------------------- Helpers ----------------------

func parseLevel(level string) zerolog.Level {
	switch level {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// ---------------------- Context helpers ----------------------

// WithTraceID adds a trace ID to the context
func WithTraceID(ctx context.Context, traceID uuid.UUID) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceIDFromContext extracts the trace ID from the context
func TraceIDFromContext(ctx context.Context) (string, bool) {
	if v := ctx.Value(traceIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

// SubLogsFromContext retrieves the captured sub logs from the context
func SubLogsFromContext(ctx context.Context) []SubLogRequest {
	if l := FromContext(ctx); l != nil {
		return l.Logs()
	}
	return nil
}

// NewContext initializes a context with a new EventLogger and trace ID
func NewContext(ctx context.Context, base *zerolog.Logger, traceID uuid.UUID) context.Context {
	logger := NewEventLogger(ctx, base, traceID)
	return WithLogger(ctx, logger)
}
