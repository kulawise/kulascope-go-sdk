package kulascope

import (
	"fmt"
	"os"
)

type Environment string

const (
	Staging    Environment = "staging"
	Production Environment = "production"
)

type Config struct {
	Environment        Environment
	APIKey             string
	RedactHeaders      []string
	RedactRequestBody  []string
	RedactResponseBody []string
	WorkerCount        int
}

func NewConfig(env string) (Config, error) {
	environment, err := ParseEnvironment(env)
	if err != nil {
		return Config{}, err
	}

	apiKey := os.Getenv("KULASCOPE_API_KEY")
	if apiKey == "" {
		return Config{}, fmt.Errorf("missing API key: set KULASCOPE_API_KEY in env")
	}

	return Config{
		Environment: environment,
		APIKey:      apiKey,
	}, nil
}
