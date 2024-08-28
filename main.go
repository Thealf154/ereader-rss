package main

import (
	"html/template"
	"os"

	"ereader-rss/app/middleware"
	"ereader-rss/app/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
)

func main() {
	engine := html.New("./app/views", ".html")
	engine.AddFunc(
		"unescape", func(s string) template.HTML {
			return template.HTML(s)
		},
	)

	app := fiber.New(fiber.Config{
		Views:        engine,
		ErrorHandler: middleware.HandleError,
	})

	app.Static("/", "./app/public/")

	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))
	app.Use(helmet.New())
	app.Use(cache.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	if os.Getenv("DEBUG") == "true" {
		log.Debug("log level set to DEBUG")
		log.SetLevel(log.LevelDebug)
	} else {
		log.SetLevel(log.LevelError)
	}

	web := app.Group("")
	routes.RegisterWeb(web)

	// Start the server on port 3000
	log.Fatal(app.Listen(":3000"))
}
