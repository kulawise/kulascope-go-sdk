package kulascope

import (
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kulawise/kulascope-go-sdk/log"
	"github.com/rs/zerolog"
)

func Middleware(cfg Config) fiber.Handler {
	Init(cfg)

	baseLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return func(c *fiber.Ctx) error {
		start := time.Now()
		traceID := uuid.New()

		rw := WrapResponseWriter(c)

		reqHeaders := make(map[string][]string)
		c.Request().Header.VisitAll(func(k, v []byte) {
			reqHeaders[string(k)] = []string{string(v)}
		})
		redactedReqHeaders := redactHeaders(reqHeaders, cfg.RedactHeaders)

		reqBody := c.Body()
		redactedReqBody := redactJSON(reqBody, cfg.RedactRequestBody)

		ctx := log.NewContext(c.UserContext(), &baseLogger, traceID)
		c.SetUserContext(ctx)

		err := c.Next()

		resHeaders := make(map[string][]string)
		c.Response().Header.VisitAll(func(k, v []byte) {
			resHeaders[string(k)] = []string{string(v)}
		})
		redactedRespHeaders := redactHeaders(resHeaders, cfg.RedactHeaders)

		respBody := c.Response().Body()
		respCopy := append([]byte(nil), respBody...)
		redactedRespBody := redactJSON(respCopy, cfg.RedactResponseBody)

		userAgent := c.Get("User-Agent")
		ip := c.Get("X-Forwarded-For")
		if ip != "" {
			ips := strings.Split(ip, ",")
			ip = strings.TrimSpace(ips[0])
		} else {
			ip = c.Get("X-Real-IP")
			if ip == "" {
				ip = c.IP()
			}
		}

		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()
		latency := int(time.Since(start).Milliseconds())

		subLogs := log.SubLogsFromContext(ctx)

		metadata := map[string]any{
			"user_agent":       userAgent,
			"request_headers":  redactedReqHeaders,
			"request_body":     string(redactedReqBody),
			"response_headers": redactedRespHeaders,
			"response_body":    string(redactedRespBody),
			"content_type":     string(c.Request().Header.ContentType()),
			"request_size":     len(reqBody),
			"response_size":    rw.Size(),
			"referer":          c.Get("Referer"),
			"host":             string(c.Request().Host()),
		}

		req := CreateLogRequest{
			TraceID:   traceID,
			Level:     "info",
			Message:   "http request completed",
			Status:    &status,
			Method:    &method,
			Path:      &path,
			Latency:   &latency,
			IP:        &ip,
			Metadata:  metadata,
			SubLogs:   subLogs,
			Timestamp: time.Now(),
		}

		sendQueue <- sendJob{cfg: cfg, payload: req}

		return err
	}
}
