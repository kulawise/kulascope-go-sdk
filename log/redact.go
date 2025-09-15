package log

import "strings"

var defaultRedactBodyKeys = []string{"password", "secret", "token"}

func redactMetadata(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}

	out := make(map[string]any, len(m))

	for k, v := range m {
		lowerKey := strings.ToLower(k)

		// If the key itself is sensitive, redact regardless of value type
		if find(defaultRedactBodyKeys, lowerKey) {
			out[k] = "[CLIENT_REDACTED]"
			continue
		}

		// Recurse if value is a map
		switch val := v.(type) {
		case map[string]any:
			out[k] = redactMetadata(val)
		case []any:
			arr := make([]any, len(val))
			for i, elem := range val {
				if nestedMap, ok := elem.(map[string]any); ok {
					arr[i] = redactMetadata(nestedMap)
				} else {
					arr[i] = elem
				}
			}
			out[k] = arr
		default:
			out[k] = val
		}
	}

	return out
}

func find(haystack []string, needle string) bool {
	for _, hay := range haystack {
		if strings.EqualFold(hay, needle) {
			return true
		}
	}
	return false
}
