package domain

import "time"

type RepositoryMember struct {
	Id           string
	RepositoryId string
	Repository   Repository `gorm:"foreignKey:RepositoryId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	MemberType MemberType

	UserId *string
	User   User `gorm:"foreignKey:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	GroupId *string
	Group   Group `gorm:"foreignKey:GroupId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	AccessType AccessType

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (RepositoryMember) TableName() string {
	return "repository_member"
}

type RepositoryMemberPers interface {
	GetByRepositoryWithRelations(repositoryId string) ([]RepositoryMember, error)
	Create(member *RepositoryMember) error
}
