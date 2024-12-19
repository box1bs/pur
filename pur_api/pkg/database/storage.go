package database

import (
	"log"

	"github.com/box1bs/pur/pur_api/pkg/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Postgres struct {
	DB *gorm.DB
}

func(p *Postgres) InitMigrate() error {
	if err := p.DB.AutoMigrate(&model.Account{}, &model.Link{}); err != nil {
		log.Printf("migration failed: %v\n", err)
		return err
	}
	
	return nil
}

func(p *Postgres) CreateAccount(account model.Account) error {
	if err := p.DB.Create(&account).Error; err != nil {
		log.Printf("failed to create account: %v\n", err)
		return err
	}

	return nil
}

func(p *Postgres) DeleteAccount(id uuid.UUID) error {
	if err := p.DB.Delete(&model.Account{},"id = ?", id).Error; err != nil  {
		log.Printf("failed to delete account: %v\n", err)
		return err
	}

	return nil
}

func(p *Postgres) GetAccountByID(id uuid.UUID) (model.Account, error) {
	var account model.Account
	if err := p.DB.First(&account,"id = ?", id).Error; err != nil {
		log.Printf("failed to get account: %v\n", err)
		return model.Account{}, err
	}

	return account, nil
}

func(p *Postgres) SaveLink(link model.Link) error {
	if err := p.DB.Create(&link).Error; err != nil {
		log.Printf("failed saving link: %v", err)
		return err
	}

	return nil
}

func(p *Postgres) UpdateLink(updatedLink model.Link) error {
	if err := p.DB.Save(&updatedLink).Error; err != nil {
		log.Printf("failed updating link: %v", err)
		return err
	}

	return nil
}

func(p *Postgres) DeleteLinkByID(id uuid.UUID) error {
	if err := p.DB.Delete(&model.Link{},"id = ?", id).Error; err != nil {
		log.Printf("failed deleting link: %v", err)
		return err
	}

	return nil
}

func(p *Postgres) DeleteAllLinksById(id uuid.UUID) error {
	if err := p.DB.Where("account_id = ?", id).Delete(&model.Link{}).Error; err != nil {
		log.Printf("failed deleting links: %v", err)
		return err
	}

	return nil
}

func(p *Postgres) GetLinksByAccountID(accountId uuid.UUID) ([]model.Link, error) {
	var links []model.Link
	if err := p.DB.Where("account_id = ?", accountId).Find(&links).Error; err != nil {
		log.Printf("failed getting links: %v", err)
		return nil, err
	}

	return links, nil
}