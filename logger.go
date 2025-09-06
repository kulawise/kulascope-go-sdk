package kulascope

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/google/uuid"
	// "github.com/kulawise/kulascope-sdk-go/log"
	"github.com/rs/zerolog"
)

type contextKey string

const traceIDKey contextKey = "trace_id"

var (
	baseLogger zerolog.Logger
	logChan    chan func(zerolog.Logger)
)

var defaultSensitiveKeys = []string{
	"password", "pass", "pwd",
	"token", "access_token", "refresh_token", "id_token",
	"authorization", "x-authorization", "x-auth-token",
	"secret", "client_secret", "private_key", "ssh_key",
	"api_key", "apikey", "x-api-key",
	"credit_card", "cc", "cvv", "pin",
}

type redactingWriter struct {
	w    io.Writer
	rexs []*regexp.Regexp
}

func NewRedactingWriter(w io.Writer, keys []string) io.Writer {
	if len(keys) == 0 {
		keys = defaultSensitiveKeys
	}
	rexs := make([]*regexp.Regexp, 0, len(keys))
	for _, k := range keys {

		p := fmt.Sprintf(`(?i)("(%s)"\\s*:\\s*)"(?:[^"\\\\]|\\\\.)*"`, regexp.QuoteMeta(k))
		rexs = append(rexs, regexp.MustCompile(p))
	}
	return &redactingWriter{w: w, rexs: rexs}
}

func (rw *redactingWriter) Write(p []byte) (int, error) {
	s := string(p)
	for _, rx := range rw.rexs {
		s = rx.ReplaceAllString(s, `$1"[REDACTED]"`)
	}
	return rw.w.Write([]byte(s))
}

func Init(cfg Config) {
	zerolog.DurationFieldUnit = time.Millisecond

	w := NewRedactingWriter(os.Stdout, cfg.RedactHeaders)

	baseLogger = zerolog.New(zerolog.SyncWriter(w)).With().
		Timestamp().
		Str("env", string(cfg.Environment)).
		Logger()

	logChan = make(chan func(zerolog.Logger), 100_000)
	startLogWorker()

	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 4 // safe default
	}
	startSenderWorkers(cfg.WorkerCount)

}

func startLogWorker() {
	go func() {
		batch := make([]func(zerolog.Logger), 0, 100) // batch size = 100
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case f := <-logChan:
				batch = append(batch, f)
				if len(batch) >= cap(batch) {
					flush(batch)
					batch = batch[:0]
				}
			case <-ticker.C:
				if len(batch) > 0 {
					flush(batch)
					batch = batch[:0]
				}
			}
		}
	}()
}

func flush(batch []func(zerolog.Logger)) {
	for _, f := range batch {
		f(baseLogger)
	}
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		traceID = uuid.New().String()
	}
	return context.WithValue(ctx, traceIDKey, traceID)
}

func FromContext(ctx context.Context) zerolog.Logger {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok && traceID != "" {
		return baseLogger.With().Str("trace_id", traceID).Logger()
	}
	return baseLogger
}

func GetTraceID(ctx context.Context) string {
	if reqID, ok := ctx.Value(traceIDKey).(string); ok && reqID != "" {
		return reqID
	}
	return ""
}

func AsyncLog(f func(zerolog.Logger)) {
	select {
	case logChan <- f:
	default:
		select {
		case <-logChan:
		default:
		}
		logChan <- f
	}
}

func AccessLog(ctx context.Context, method, path string, status int, start time.Time, ip string) {
	duration := time.Since(start)
	traceLogger := FromContext(ctx)

	AsyncLog(func(log zerolog.Logger) {
		traceLogger.Info().
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Dur("latency", duration).
			Str("ip", ip).
			Msg("http request completed")
	})
}

// const logsKey contextKey = "request_logs"

// func getLogs(ctx context.Context) []log.SubLogRequest {
// 	if logs, ok := ctx.Value(logsKey).([]log.SubLogRequest); ok {
// 		return logs
// 	}
// 	return nil
// }
