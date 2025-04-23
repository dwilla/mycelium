package handlers

import "github.com/dwilla/mycelium/internal/database"

type Config struct {
	DB          *database.Queries
	CurrentUser database.User
}
