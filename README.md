# kulascope-go-sdk
This SDK is written to be used in a Go Fiber server, for observability

## Installation

```
go get github.com/kulawise/kulascope-go-sdk
```


## Usage
```
package main

import (
    "log"
    "time"

    kulascope "github.com/kulawise/kulascope-go-sdk"
    logger "github.com/kulawise/kulascope-go-sdk/log"
    "github.com/gofiber/fiber/v2"
)

func main() {
    ksCfg := kulascope.Config{
		Environment:        kulascope.Staging,
		APIKey:             "your_api_key",
		RedactHeaders:      []string{"Authorization"},
		RedactRequestBody:  []string{"$.password"},
		RedactResponseBody: []string{"$.password"},
	}

    app := fiber.New()
    app.Use(kulascope.Middleware(ksCfg))

    app.Get("/", func(c *fiber.Ctx) error {
        log := logger.WithContext(c.UserContext())
        log.Info("Hello route called", nil)
        return c.SendString("Hello, world!")
    })

    app.Listen(":8080")
}
```
