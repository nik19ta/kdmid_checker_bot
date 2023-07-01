package models

import (
	"github.com/google/uuid"
)

type Request struct {
	ID                uuid.UUID `gorm:"type:uuid;"`        // * ID
	UserID            int64     `gorm:"type:integer"`      // * Id юзера в телеграмме
	ApplicationNumber string    `gorm:"type:integer"`      // * Номер заявления
	CityID            int       `gorm:"type:integer"`      // * Id города
	NumberChecksToday int       `gorm:"type:integer"`      // * Предыдущий статус заявления
	PassportType      string    `gorm:"type:varchar(255)"` // * Тип паспорта, 10 или 5 лет
	Status            string    `gorm:"type:varchar(255)"` // * Предыдущий статус заявления
}
