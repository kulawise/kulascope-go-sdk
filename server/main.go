package main

import (
	l "log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kulawise/kulascope-sdk-go"
	logger "github.com/kulawise/kulascope-sdk-go/log"
)

func main() {
	app := fiber.New()

	cfg := kulascope.Config{
		Environment:        kulascope.Staging,
		APIKey:             "ks_bf71a9ea.UOBzx0u06kDr3iQQtVsgQoBRn0enbRBHRqc2ATYuuN8",
		RedactHeaders:      []string{"Authorization"},
		RedactRequestBody:  []string{"$.password"},
		RedactResponseBody: []string{"$.password"},
		WorkerCount:        2,
	}

	app.Use(kulascope.Middleware(cfg))

	app.Get("/login", func(c *fiber.Ctx) error {
		log := logger.FromContext(c.UserContext())

		log.Info().
			Str("user_id", uuid.NewString()).
			Int("attempts", 3).
			Msg("User authenticated successfully")

		log.Debug().
			Msg("Checking credentials against database")

		log.Error().
			Msg("User authenticated successfully")

		log.Warn().
			Msg("Something went wrong")

		return c.SendString("ok")
	})

	l.Println("Server running on :3005")
	app.Listen(":3005")
}
