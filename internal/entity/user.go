package entity

import "time"

type User struct {
	ID           string
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}
