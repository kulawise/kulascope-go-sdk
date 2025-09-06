package log

import (
	"context"
	"sync"

	"github.com/rs/zerolog"
)

type contextKey string

const logsKey contextKey = "request_logs"

// SubLogRequest represents an individual log entry stored per request
type SubLogRequest struct {
	Level    string         `json:"level"`
	Message  string         `json:"message"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

var (
	mu        sync.RWMutex
	logger    *zerolog.Logger
	hasLogger bool

	// current request context, set by middleware per request
	requestCtx context.Context
)

// SetLogger sets the global zerolog logger
func SetLogger(l zerolog.Logger) {
	mu.Lock()
	defer mu.Unlock()
	logger = &l
	hasLogger = true
}

// SetContext sets the current request context (middleware must call this per request)
func SetContext(ctx context.Context) {
	requestCtx = ctx
}

// internal helper to get the logger safely
func getLogger() *zerolog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	if !hasLogger {
		nop := zerolog.Nop()
		return &nop
	}
	return logger
}

// appendLog adds a log entry to the request context
func appendLog(ctx context.Context, level, message string, metadata map[string]any) context.Context {
	logs, _ := ctx.Value(logsKey).([]SubLogRequest)
	logs = append(logs, SubLogRequest{
		Level:    level,
		Message:  message,
		Metadata: metadata,
	})
	return context.WithValue(ctx, logsKey, logs)
}

// GetLogs returns all sublogs for the current context
func GetLogs(ctx context.Context) []SubLogRequest {
	if logs, ok := ctx.Value(logsKey).([]SubLogRequest); ok {
		return logs
	}
	return nil
}

// --- Logging functions that automatically use the requestCtx ---

func Info(msg string, metadata map[string]any) {
	if requestCtx != nil {
		requestCtx = appendLog(requestCtx, "info", msg, metadata)
	}
	getLogger().Info().Fields(metadata).Msg(msg)
}

func Error(msg string, metadata map[string]any) {
	if requestCtx != nil {
		requestCtx = appendLog(requestCtx, "error", msg, metadata)
	}
	getLogger().Error().Fields(metadata).Msg(msg)
}

func Warn(msg string, metadata map[string]any) {
	if requestCtx != nil {
		requestCtx = appendLog(requestCtx, "warn", msg, metadata)
	}
	getLogger().Warn().Fields(metadata).Msg(msg)
}

func Debug(msg string, metadata map[string]any) {
	if requestCtx != nil {
		requestCtx = appendLog(requestCtx, "debug", msg, metadata)
	}
	getLogger().Debug().Fields(metadata).Msg(msg)
}

func Trace(msg string, metadata map[string]any) {
	if requestCtx != nil {
		requestCtx = appendLog(requestCtx, "trace", msg, metadata)
	}
	getLogger().Trace().Fields(metadata).Msg(msg)
}

func Fatal(msg string, metadata map[string]any) {
	if requestCtx != nil {
		requestCtx = appendLog(requestCtx, "fatal", msg, metadata)
	}
	getLogger().Fatal().Fields(metadata).Msg(msg)
}

func Panic(msg string, metadata map[string]any) {
	if requestCtx != nil {
		requestCtx = appendLog(requestCtx, "panic", msg, metadata)
	}
	getLogger().Panic().Fields(metadata).Msg(msg)
}
