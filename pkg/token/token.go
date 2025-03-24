package token

import (
	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"time"
)

// Maker is an interface for managing tokens
type Maker interface {
	// CreateToken creates a new token for a specific username and duration
	CreateToken(username string, role db.Role, duration time.Duration, userID int64) (string, *Payload, error)

	// VerifyToken checks if the token is valid or not
	VerifyToken(token string) (*Payload, error)
}
