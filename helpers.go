package kulascope

import (
	"fmt"
	"strings"
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
