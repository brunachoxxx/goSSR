package routes

import (
	"goSSR/auth"
	"goSSR/handlers"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	h := handlers.NewHandler(db)
	authHandler := auth.NewHandler(db)

	app.Get("/", h.HandleIndex)
	app.Get("/about", h.HandleAbout)
	app.Post("/upload", auth.RequireAuth, h.HandleUpload)
	app.Get("/auth/google", authHandler.GoogleLoginHandler)
	app.Get("/auth/google/callback", authHandler.GoogleCallbackHandler)
	app.Get("/logout", authHandler.HandleLogout)
	app.Post("/delete/:id", h.HandleDeleteImage)
	app.Get("/logout-dialog", h.HandleLogoutDialog)
	app.Post("/delete-dialog/:id", h.HandleDeleteImageDialog)
	app.Get("polaroid", h.HandlePolaroid)
	app.Get("/cookies", h.HandleCookies)
	// New SEO-related routes
	app.Get("/sitemap.xml", h.HandleSitemap)
	app.Get("/robots.txt", h.HandleRobots)
}
