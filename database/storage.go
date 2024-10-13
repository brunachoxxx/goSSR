package database

import (
	"os"
	"time"

	"github.com/gofiber/storage/postgres/v3"
)

type Config = postgres.Config

// Config for storage
var ConfigStorage = Config{
	ConnectionURI: os.Getenv("STORAGE_DB_URL"),
	Reset:         false,
	GCInterval:    10 * time.Second,
}

// config for databse
