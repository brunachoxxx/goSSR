package main

import (
	"github.com/gofiber/fiber/v2"
)

func CustomErrorHandler(c *fiber.Ctx, err error) error {
	// Default 500 status code for internal server errors
	code := fiber.StatusInternalServerError

	// If it's a Fiber-specific error, extract the error code
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Set the status code for the response
	c.Status(code)

	data := fiber.Map{
		"PageTitle":   "Error",
		"Title":       "An Error Occurred",
		"Description": "We encountered an error while processing your request.",
		"Error":       err.Error(),
		"StatusCode":  code,
	}

	// Add NavItems if you want to show navigation on the error page
	// data["NavItems"] = ... // Add your navigation items here

	// Render the error page with the layout, passing the error message
	return c.Render("error", data, "layouts/main")
}
