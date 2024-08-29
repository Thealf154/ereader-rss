package middleware

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func HandleError(ctx *fiber.Ctx, err error) error {
	// Status code defaults to 500
	code := fiber.StatusInternalServerError

	// Retrieve the custom status code if it's a *fiber.Error
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	log.Error(fmt.Errorf("exception: %s ocurred in url: %s ", ctx.Query("url"), err))

	// Send custom error page
	err = ctx.Status(code).Render("error", fiber.Map{
		"Error": fmt.Errorf("something wrong happened"),
	})

	if err != nil {
		// In case the SendFile fails
		return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Return from handler
	return nil
}
