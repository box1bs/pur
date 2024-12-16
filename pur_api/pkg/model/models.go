package model

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	Id 			uuid.UUID   `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"user_id"`
	Name 		string		`json:"name"`
	CreatedAt 	time.Time	`gorm:"default:current_timestamp" json:"created_at"`
}

type Link struct {
	Id 			uuid.UUID 	`gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Url			string		`json:"url"`
	Description string		`json:"description"`
	Summary		string		`json:"summary"`
	Type		string		`json:"type"`
	AccountId	uuid.UUID	`json:"user_id" gorm:"foreignKey:AccountId;references:Id"`
	SavedAt		time.Time	`json:"saved_at" gorm:"default:current_timestamp"`
	State		bool		`json:"state"` //used or not
}