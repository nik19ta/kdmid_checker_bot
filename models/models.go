package models

import (
	"github.com/google/uuid"
)

type Request struct {
	ID                uuid.UUID `gorm:"type:uuid;"`
	UserID            int64     `gorm:"type:integer"`
	ApplicationNumber string    `gorm:"type:integer"`
	CityID            int       `gorm:"type:integer"`
	PassportType      string    `gorm:"type:varchar(255)"`
}
