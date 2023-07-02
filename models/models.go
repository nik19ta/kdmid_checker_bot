package models

import (
	"github.com/google/uuid"
)

type Request struct {
	ID                uuid.UUID `gorm:"type:uuid;"`        // * ID
	UserID            int64     `gorm:"type:integer"`      // * Telegram User ID
	ApplicationNumber string    `gorm:"type:integer"`      // * Application Number
	CityID            int       `gorm:"type:integer"`      // * City ID
	NumberChecksToday int       `gorm:"type:integer"`      // * Previous Application Status
	PassportType      string    `gorm:"type:varchar(255)"` // * Passport Type (10 or 5 years)
	Status            string    `gorm:"type:varchar(255)"` // * Previous Application Status
}
