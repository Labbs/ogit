package persistence

import (
	"github.com/labbs/ogit/domain"
	"gorm.io/gorm"
)

type repositoryMemberPers struct {
	db *gorm.DB
}

func NewRepositoryMemberPers(db *gorm.DB) *repositoryMemberPers {
	return &repositoryMemberPers{db: db}
}

func (r *repositoryMemberPers) GetByRepositoryWithRelations(repositoryId string) ([]domain.RepositoryMember, error) {
	var members []domain.RepositoryMember
	err := r.db.Preload("User").Preload("Group").Where("repository_id = ?", repositoryId).Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (r *repositoryMemberPers) Create(member *domain.RepositoryMember) error {
	return r.db.Create(member).Error
}
