package persistence

import (
	"github.com/labbs/ogit/domain"
	"gorm.io/gorm"
)

type repositoryPers struct {
	db *gorm.DB
}

func NewRepositoryPers(db *gorm.DB) *repositoryPers {
	return &repositoryPers{db: db}
}

func (r *repositoryPers) GetRepository(repo domain.Repository) (*domain.Repository, error) {
	var doc domain.Repository
	err := r.db.Where(repo).First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *repositoryPers) Create(repo *domain.Repository) error {
	return r.db.Debug().Create(repo).Error
}

func (r *repositoryPers) Delete(repo domain.Repository) error {
	return r.db.Where("id = ?", repo.Id).Delete(&domain.Repository{}).Error
}
