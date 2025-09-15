package kulascope

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
)

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
func RedactJSON(data []byte, redactList []string) []byte {
	if len(data) == 0 {
		return data
	}

	var src interface{}
	if err := json.Unmarshal(data, &src); err != nil {
		// if not JSON, return as-is
		return data
	}

	RedactRecursive(src, redactList)

	out, err := json.Marshal(src)
	if err != nil {
		return data
	}
	return out
}

// helper: recursively walk JSON object
func RedactRecursive(node interface{}, redactList []string) {
	switch v := node.(type) {
	case map[string]interface{}:
		for key, val := range v {
			// if the key matches the redact rules, replace the value
			if keyMatchesRedact(key, redactList) {
				v[key] = "[CLIENT_REDACTED]"
				continue
			}

			// If the value is a string that *looks like* JSON, try to parse and redact inside it
			if s, ok := val.(string); ok {
				ts := strings.TrimSpace(s)
				if len(ts) > 0 && (ts[0] == '{' || ts[0] == '[') {
					var nested interface{}
					if err := json.Unmarshal([]byte(ts), &nested); err == nil {
						RedactRecursive(nested, redactList)
						if b, err := json.Marshal(nested); err == nil {
							v[key] = string(b)
							continue
						}
					}
				}
			}

			// Otherwise recurse normally
			RedactRecursive(val, redactList)
		}

	case []interface{}:
		for i := range v {
			RedactRecursive(v[i], redactList)
		}
	}
}

// RedactHeaders replaces sensitive headers with "[CLIENT_REDACTED]"
func redactHeaders(headers map[string][]string, redactList []string) map[string][]string {
	for k := range headers {
		if keyMatchesRedact(k, redactList) {
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

func mergeRedactKeys(defaults, custom []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(defaults)+len(custom))
	for _, k := range append(defaults, custom...) {
		s := strings.ToLower(strings.TrimSpace(k))
		if s == "" {
			continue
		}
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}

func keyMatchesRedact(key string, redactList []string) bool {
	keyLower := strings.ToLower(strings.TrimSpace(key))
	for _, r := range redactList {
		if r == "" {
			continue
		}
		// redactList entries are expected to be lowercased (see mergeRedactKeys)
		if keyLower == r || strings.Contains(keyLower, r) || strings.Contains(r, keyLower) {
			return true
		}
	}
	return false
}
