package main

import (
	"github.com/gofiber/fiber/v2"
)

func CustomeErrorHandler(c *fiber.Ctx, err error) error {
	// Default 500 status code for internal server errors
	code := fiber.StatusInternalServerError

	// If it's a Fiber-specific error, extract the error code
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Set the status code for the response
	c.Status(code)

	// Render the error page with the layout, passing the error message
	return c.Render("error", fiber.Map{
		"Error": err.Error(),
	}, "layouts/main")
}
