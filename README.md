# kulascope-go-sdk
This SDK is written to be used in a Go Fiber server, for observability


## Usage
```
package main

import (
    "log"
    "time"

    "github.com/kulawise/kulascope-sdk-go"
    appLog "github.com/kulawise/kulascope-sdk-go/log"
    "github.com/gofiber/fiber/v2"
)

func main() {
    cfg, err := kulascope.NewConfig("staging")
    if err != nil {
        log.Fatal(err)
    }

    app := fiber.New()
    app.Use(kulascope.Middleware(cfg))

    app.Get("/", func(c *fiber.Ctx) error {
        l := appLog.WithContext(c.UserContext())
        l.Info("Hello route called", nil)
        return c.SendString("Hello, world!")
    })

    app.Listen(":8080")
}
```
