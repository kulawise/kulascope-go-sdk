package kulascope

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kulawise/kulascope-sdk-go/log"
	"github.com/rs/zerolog"
)

func Middleware(cfg Config) fiber.Handler {
	Init(cfg)

	// Create the base zerolog logger internally
	baseLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return func(c *fiber.Ctx) error {
		start := time.Now()
		traceID := uuid.New()

		// Create a new request context with logger + trace ID
		ctx := log.NewContext(c.UserContext(), &baseLogger, traceID)
		c.SetUserContext(ctx)

		// Call next handler
		err := c.Next()

		// Collect request metadata
		userAgent := c.Get("User-Agent")
		ip := c.IP()
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()
		latency := int(time.Since(start).Milliseconds())

		// Collect sub logs from this request
		subLogs := log.SubLogsFromContext(ctx)

		// Create top-level request log

		req := CreateLogRequest{
			TraceID: traceID,
			Level:   "info",
			Message: "http request completed",
			Status:  &status,
			Method:  &method,
			Path:    &path,
			Latency: &latency,
			IP:      &ip,
			Metadata: map[string]any{
				"user_agent": userAgent,
			},
			SubLogs: subLogs,
		}

		// Enqueue log for asynchronous sending
		sendQueue <- sendJob{cfg: cfg, payload: req}

		return err
	}
}
