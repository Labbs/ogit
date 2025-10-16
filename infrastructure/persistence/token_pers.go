package persistence

import (
	"github.com/labbs/ogit/domain"
	"gorm.io/gorm"
)

type tokenPers struct {
	db *gorm.DB
}

func NewTokenPers(db *gorm.DB) *tokenPers {
	return &tokenPers{db: db}
}

func (t *tokenPers) Create(token *domain.Token) error {
	return t.db.Create(&token).Error
}

func (t *tokenPers) GetByNameAndToken(name, tokenStr string) (*domain.Token, error) {
	var token domain.Token
	err := t.db.Preload("User").Preload("Repository").Where("name = ? AND token = ?", name, tokenStr).First(&token).Error
	return &token, err
}

func (t *tokenPers) GetByUserId(userId string) ([]domain.Token, error) {
	var tokens []domain.Token
	err := t.db.Preload("User").Preload("Repository").Where("user_id = ?", userId).Find(&tokens).Error
	return tokens, err
}

func (t *tokenPers) GetById(id string) (*domain.Token, error) {
	var token domain.Token
	err := t.db.Preload("User").Preload("Repository").Where("id = ?", id).First(&token).Error
	return &token, err
}

func (t *tokenPers) Delete(token *domain.Token) error {
	return t.db.Where("id = ? AND user_id = ?", token.Id, token.UserId).Delete(&domain.Token{}).Error
}
