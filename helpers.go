package kulascope

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func ParseEnvironment(s string) (Environment, error) {
	switch strings.ToLower(s) {
	case "staging":
		return Staging, nil
	case "production", "prod":
		return Production, nil
	default:
		return "", fmt.Errorf("invalid environment: %s (must be 'staging' or 'production')", s)
	}
}

func BuildCreateLogRequest(ctx context.Context, traceID uuid.UUID, level string, message string, status int, method, path, ip string, start time.Time) CreateLogRequest {
	latency := int(time.Since(start).Milliseconds())

	return CreateLogRequest{
		TraceID:  traceID,
		Level:    level,
		Message:  message,
		Metadata: map[string]any{},
		Status:   &status,
		Method:   &method,
		Path:     &path,
		Latency:  &latency,
		IP:       &ip,
		SubLogs:  getLogs(ctx),
	}
}
