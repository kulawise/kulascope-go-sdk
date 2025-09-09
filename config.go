package kulascope

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
