package routes

import (
	Controller "ereader-rss/app/controllers"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func RegisterWeb(web fiber.Router) {
	web.Get("/rss/list", Controller.GetListFromRSS())
	web.Get("/rss/page", Controller.GetPageFromRSS())
	web.Get("/rss/epub/download.epub", Controller.GetRssAsEpub())

	web.Get("/health", func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) })
	web.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Render("home", fiber.Map{})
	})
}
