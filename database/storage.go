package database

import (
	"time"

	"github.com/gofiber/storage/postgres/v3"
)

type Config = postgres.Config

// Config for storage
var ConfigStorage = Config{
	ConnectionURI: "postgresql://admin:password@localhost:5432/gossr",
	Host:          "127.0.0.1",
	Port:          5432,
	Database:      "gossr",
	Table:         "gossr_storage",
	SSLMode:       "disable",
	Reset:         false,
	GCInterval:    10 * time.Second,
}

// config for databse
