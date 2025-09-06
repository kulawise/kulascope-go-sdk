package kulascope

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kulawise/kulascope-sdk-go/log"
)

func Middleware(cfg Config) fiber.Handler {
	Init(cfg)

	return func(c *fiber.Ctx) error {
		// generate trace ID
		traceID := uuid.New()
		start := time.Now()

		// create request context with traceID and empty sublogs
		ctx := context.WithValue(c.UserContext(), traceIDKey, traceID.String())
		ctx = context.WithValue(ctx, logsKey, []SubLogRequest{})
		c.SetUserContext(ctx)

		// tell the log package about the current request context
		log.SetContext(ctx)

		// call next handler
		err := c.Next()

		// record the access log as a sublog
		AccessLog(
			c.UserContext(),
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			start,
			c.IP(),
		)

		// build the payload for analysis server
		req := BuildCreateLogRequest(
			c.UserContext(),
			traceID,
			"info",
			"http request completed",
			c.Response().StatusCode(),
			c.Method(),
			c.Path(),
			c.IP(),
			start,
		)

		// enqueue the log for sending asynchronously
		sendQueue <- sendJob{cfg: cfg, payload: req}

		return err
	}
}
