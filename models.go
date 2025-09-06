package kulascope

import (
	"github.com/google/uuid"
	"github.com/kulawise/kulascope-sdk-go/log"
)

type CreateLogRequest struct {
	TraceID  uuid.UUID           `json:"trace_id"`
	Level    string              `json:"level"`
	Message  string              `json:"message"`
	Metadata map[string]any      `json:"metadata"`
	Status   *int                `json:"status,omitempty"`
	Method   *string             `json:"method,omitempty"`
	Path     *string             `json:"path,omitempty"`
	Latency  *int                `json:"latency,omitempty"`
	IP       *string             `json:"ip,omitempty"`
	SubLogs  []log.SubLogRequest `json:"sub_logs"`
}
