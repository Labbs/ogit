package persistence

import (
	"github.com/labbs/ogit/domain"
	"gorm.io/gorm"
)

type sessionPers struct {
	db *gorm.DB
}

func NewSessionPers(db *gorm.DB) *sessionPers {
	return &sessionPers{db: db}
}

func (s *sessionPers) Create(session *domain.Session) error {
	return s.db.Create(session).Error
}

func (s *sessionPers) GetById(id string) (*domain.Session, error) {
	var session domain.Session
	err := s.db.Where("id = ?", id).First(&session).Error
	return &session, err
}
