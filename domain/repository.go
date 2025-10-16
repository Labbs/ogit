package domain

import "time"

type Repository struct {
	Id          string `gorm:"primaryKey"`
	Name        string
	Description string
	Slug        string

	DefaultBranch string

	IsArchived bool `gorm:"default:false"`
	ArchivedAt time.Time

	IsPrivate bool `gorm:"default:true"`

	// DeployKeys DeployKeys `gorm:"foreignKey:RepositoryId"`
	// WebHooks   WebHooks   `gorm:"foreignKey:RepositoryId"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Repository) TableName() string {
	return "repository"
}

// DeployKeys represents a list of deploy keys for a repository.
type DeployKeys []DeployKey

type DeployKey struct {
	Name       string
	Key        string
	AccessType AccessType
	CreatedAt  time.Time
}

// WebHook represents a webhook for a repository.
type WebHooks []WebHook
type WebHook struct {
	Id          string
	Url         string
	ContentType ContentType
	Secret      string
	SSLVerify   bool
	Events      Events
	Actived     bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

type ContentType string
type JsonContentType ContentType

type Events []Event

type Event string
type AllEvents Event
type PushEvent Event
type PullRequestEvent Event
type IssuesEvent Event
type IssueCommentEvent Event

type RepositoryPers interface {
	GetRepository(repo Repository) (*Repository, error)
	Create(repo *Repository) error
	Delete(repo Repository) error
}
