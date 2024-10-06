// goSSR/handlers/seo_handlers.go

package handlers

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/ikeikeikeike/go-sitemap-generator/v2/stm"
)

// SEO-related handlers for sitemap and robots.txt

func (h *Handler) HandleSitemap(c *fiber.Ctx) error {
	domain := os.Getenv("DOMAIN") // Make sure to set this in your .env file
	if domain == "" {
		domain = "http://localhost:3000" // Fallback for development
	}

	sm := stm.NewSitemap(1)
	sm.SetDefaultHost(domain)

	// Add static pages
	sm.Create()
	sm.Add(stm.URL{{"loc", "/"}, {"changefreq", "daily"}, {"priority", 1.0}})
	sm.Add(stm.URL{{"loc", "/about"}, {"changefreq", "weekly"}, {"priority", 0.8}})
	sm.Add(stm.URL{{"loc", "/cookies"}, {"changefreq", "monthly"}, {"priority", 0.5}})
	// Add any other static routes you want to be indexed

	// Generate the sitemap
	c.Set("Content-Type", "application/xml")
	return c.Send(sm.XMLContent())
}

func (h *Handler) HandleRobots(c *fiber.Ctx) error {
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		domain = "http://localhost:3000" // Fallback for development
	}

	robotsTxt := `User-agent: *
Allow: /
Sitemap: ` + domain + `/sitemap.xml`

	return c.SendString(robotsTxt)
}
