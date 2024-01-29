package dal

import "time"

type (
	User struct {
		ID           uint64
		Name         string
		Password     string
		PasswordSalt string
		CreatedAt    time.Time
		UpdatedAt    time.Time
	}

	Session struct {
		SessionID uint64
		Name      string
		UserID    uint64
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	Clipboard struct {
		SessionID   uint64
		ContentType string
		Content     []byte
		UpdatedAt   time.Time
	}
)
