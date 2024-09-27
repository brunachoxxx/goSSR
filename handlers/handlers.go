package handlers

import (
	"encoding/base64"
	"goSSR/database"
	"io"
	"log"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"gorm.io/gorm"
)

type NavItem struct {
	Link      string
	Text      string
	IsCurrent bool
}

type Handler struct {
	DB *gorm.DB
}

// Variables
var pageTitle = "Simple SSR Page"
var navItems = []NavItem{
	{Link: "/", Text: "Home"},
	{Link: "/about", Text: "About"},
	// Add more navigation items as needed
}

// Handlers

// NewHandler creates a new Handler instance
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{DB: db}
}

func baseTemplateData(title, description, currentPath string) fiber.Map {
	currentNavItems := make([]NavItem, len(navItems))
	for i, item := range navItems {
		currentNavItems[i] = NavItem{
			Link:      item.Link,
			Text:      item.Text,
			IsCurrent: item.Link == currentPath,
		}
	}
	return fiber.Map{
		"PageTitle":   pageTitle,
		"Title":       title,
		"Description": description,
		"NavItems":    currentNavItems,
	}
}

// Handlers
func (h *Handler) HandleIndex(c *fiber.Ctx) error {
	log.Println("Starting HandleIndex")

	sess, ok := c.Locals("session").(*session.Session)
	if !ok {
		log.Println("Session not found in context")
		return c.Render("index", fiber.Map{"IsLoggedIn": false}, "layouts/main")
	}

	googleID := sess.Get("user_id")
	data, err := h.prepareIndexData(c, googleID)
	if err != nil {
		return err
	}

	log.Println("Rendering index page")
	return c.Render("index", data, "layouts/main")
}

func (h *Handler) HandleUpload(c *fiber.Ctx) error {
	// Get the file from the form
	file, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Failed to get file from form")
	}

	// Validate file type
	if !isValidImageType(file.Filename) {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid file type. Only JPG, JPEG, and PNG are allowed")
	}

	// Validate file size (5MB max)
	if file.Size > 5*1024*1024 {
		return fiber.NewError(fiber.StatusBadRequest, "File size exceeds 5MB limit")
	}

	src, err := file.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to open file")
	}
	defer src.Close()

	// Read the file content
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to read file")
	}

	// Encode to base64
	base64String := base64.StdEncoding.EncodeToString(fileBytes)

	// Get the current user from the session
	sess, googleID, err := GetSessionAndUserID(c)
	if err != nil {
		return err // This will automatically send a 401 Unauthorized response
	}

	// Find the user in the database
	var user database.User
	if err := h.DB.Where("google_id = ?", googleID).First(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to find user")
	}

	// Check if the image already exists for this user
	var existingImage database.Image
	err = h.DB.Where("user_google_id = ? AND base64_string = ?", googleID, base64String).First(&existingImage).Error
	if err == nil {
		// Image already exists
		return fiber.NewError(fiber.StatusConflict, "This image has already been uploaded")
	} else if err != gorm.ErrRecordNotFound {
		// Some other error occurred
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to check for existing image")
	}

	// If we get here, the image doesn't exist, so we can create a new one
	newImage := database.Image{
		UserGoogleID: googleID.(string),
		Base64String: base64String,
	}

	// Save the image to the database
	if err := h.DB.Create(&newImage).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save image to database")
	}

	data, err := h.prepareIndexData(c, googleID)
	if err != nil {
		return err
	}

	data["Success"] = "File uploaded and processed successfully"
	data["UploadedName"] = file.Filename
	data["ResetForm"] = true

	// Ensure the session is saved
	if err := sess.Save(); err != nil {
		log.Printf("Failed to save session: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save session")
	}
	log.Println("Session saved successfully")

	log.Println("Rendering index page with updated image list")
	return c.Render("index", data, "layouts/main")

}

func (h *Handler) HandleAbout(c *fiber.Ctx) error {
	data := baseTemplateData("About Us", "Learn more about our company", "/about")
	return c.Render("about", data, "layouts/main")
}

func (h *Handler) prepareIndexData(c *fiber.Ctx, googleID interface{}) (fiber.Map, error) {
	data := baseTemplateData("Home", "Welcome to our site", "/")
	data["Greeting"] = "Welcome to the homepage"
	data["ResetForm"] = false
	data["Error"] = ""
	data["IsLoggedIn"] = googleID != nil

	if googleID != nil {
		var user database.User
		if err := h.DB.Where("google_id = ?", googleID).First(&user).Error; err == nil {
			data["UserEmail"] = user.Email

			// Fetch user images
			images, err := getAllUserImages(googleID.(string), h.DB)
			if err != nil {
				log.Printf("Failed to fetch user images: %v", err)
				data["Error"] = "An error occurred while fetching user images"
			} else {
				data["UserImages"] = images
				log.Printf("Fetched %d images for user", len(images))
			}
		} else {
			log.Printf("Failed to fetch user data: %v", err)
			data["Error"] = "An error occurred while fetching user data"
		}
	}

	return data, nil
}

// Validate image file type
func isValidImageType(filename string) bool {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg", ".png":
		return true
	default:
		return false
	}
}

func getAllUserImages(googleID string, db *gorm.DB) ([]database.Image, error) {
	var images []database.Image
	err := db.Where("user_google_id = ?", googleID).Find(&images).Error
	return images, err
}

func (h *Handler) deleteSingleUserImage(googleID, base64String string) error {
	var image database.Image
	err := h.DB.Where("user_google_id = ? AND base64_string = ?", googleID, base64String).First(&image).Error
	if err != nil {
		return err
	}

	if err := h.DB.Delete(&image).Error; err != nil {
		return err
	}

	return nil
}

func (h *Handler) HandleDeleteImage(c *fiber.Ctx) error {
	imageID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid image ID"})
	}

	// Get the current user from the session
	sess, userGoogleID, err := GetSessionAndUserID(c)
	if err != nil {
		return err // This will automatically send a 401 Unauthorized response
	}

	// Delete the image
	result := h.DB.Where("id = ? AND user_google_id = ?", imageID, userGoogleID).Delete(&database.Image{})
	if result.Error != nil {
		sess.Set("flash", "Failed to delete image")
		if err := sess.Save(); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to save session")
		}
		return c.Redirect("/")
	}

	if result.RowsAffected == 0 {
		sess.Set("flash", "Image not found or not owned by user")
		if err := sess.Save(); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to save session")
		}
		return c.Redirect("/")
	}

	// Set success flash message
	sess.Set("flash", "Image deleted successfully")
	if err := sess.Save(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save session")
	}

	return c.Redirect("/")
}

func GetSessionAndUserID(c *fiber.Ctx) (*session.Session, interface{}, error) {
	sess, ok := c.Locals("session").(*session.Session)
	if !ok || sess == nil {
		return nil, nil, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: No valid session")
	}

	userGoogleID := sess.Get("user_id")
	if userGoogleID == nil {
		return nil, nil, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: User not authenticated")
	}

	return sess, userGoogleID, nil
}
