package config

import (
	"github.com/box1bs/pur/pur_api/pkg/model"
	"github.com/google/uuid"
)

type Storage interface {
	InitMigrate() error
	
	CreateAccount(model.Account) error
	DeleteAccount(uuid.UUID) error
	GetAccountByID(uuid.UUID) (model.Account, error)

	SaveLink(model.Link) error
	UpdateLink(model.Link) error
	DeleteLinkByID(uuid.UUID) error
	GetLinksByAccountID(uuid.UUID) ([]model.Link, error)
}