package kulascope

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func parseEnvironment(s string) (Environment, error) {
	switch strings.ToLower(s) {
	case "staging":
		return Staging, nil
	case "production", "prod":
		return Production, nil
	default:
		return "", fmt.Errorf("invalid environment: %s (must be 'staging' or 'production')", s)
	}
}

// find checks case-insensitive membership
func find(haystack []string, needle string) bool {
	for _, hay := range haystack {
		if strings.EqualFold(hay, needle) {
			return true
		}
	}
	return false
}

// RedactJSON walks through the JSON object and replaces values of sensitive keys
func redactJSON(data []byte, redactList []string) []byte {
	if len(data) == 0 {
		return data
	}

	var src interface{}
	if err := json.Unmarshal(data, &src); err != nil {
		// if not JSON, return as-is
		return data
	}

	redactRecursive(src, redactList)

	out, err := json.Marshal(src)
	if err != nil {
		return data
	}
	return out
}

// helper: recursively walk JSON object
func redactRecursive(node interface{}, redactList []string) {
	switch v := node.(type) {
	case map[string]interface{}:
		for key, val := range v {
			if find(redactList, key) {
				v[key] = "[CLIENT_REDACTED]"
			} else {
				redactRecursive(val, redactList)
			}
		}
	case []interface{}:
		for i := range v {
			redactRecursive(v[i], redactList)
		}
	}
}

// RedactHeaders replaces sensitive headers with "[CLIENT_REDACTED]"
func redactHeaders(headers map[string][]string, redactList []string) map[string][]string {
	for k := range headers {
		if find(redactList, k) {
			headers[k] = []string{"[CLIENT_REDACTED]"}
		}
	}
	return headers
}

type responseWriterWrapper struct {
	*fiber.Ctx
	size int
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	n := len(b)
	rw.size += n
	return rw.Ctx.Response().BodyWriter().Write(b)
}

func (rw *responseWriterWrapper) Size() int {
	return rw.size
}

// WrapResponseWriter attaches the wrapper to ctx
func WrapResponseWriter(c *fiber.Ctx) *responseWriterWrapper {
	return &responseWriterWrapper{Ctx: c, size: 0}
}
