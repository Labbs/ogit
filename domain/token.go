package domain

import (
	"database/sql/driver"
	"time"

	"github.com/goccy/go-json"

	"github.com/gofiber/fiber/v2/utils"
	"gorm.io/gorm"
)

type Token struct {
	Id          string
	Name        string
	Description string
	Token       string

	Type TokenType

	UserId *string
	User   User `gorm:"foreignKey:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	RepositoryId *string
	Repository   Repository `gorm:"foreignKey:RepositoryId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Active bool

	Scopes TokenScopes

	ExpiresAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Token) TableName() string {
	return "token"
}

func (t *Token) BeforeCreate(tx *gorm.DB) error {
	t.Id = utils.UUIDv4()
	return nil
}

type TokenType string

const (
	TokenUser TokenType = "user"
	TokenRepo TokenType = "repo"
)

type TokenScopes []TokenScope

type TokenScope string

const (
	ScopeReadRepository     TokenScope = "read:repository"
	ScopeWriteRepository    TokenScope = "write:repository"
	ScopeAdminRepository    TokenScope = "admin:repository"
	ScopeCreatePullRequest  TokenScope = "create:pull_request"
	ScopeMergePullRequest   TokenScope = "merge:pull_request"
	ScopeCommentPullRequest TokenScope = "comment:pull_request"
)

func (t TokenScopes) Value() (driver.Value, error) {
	valueString, err := json.Marshal(t)
	return string(valueString), err
}

func (t *TokenScopes) Scan(value any) error {
	return json.Unmarshal([]byte(value.(string)), t)
}

type TokenPers interface {
	Create(token *Token) error
	GetByNameAndToken(name, tokenStr string) (*Token, error)
	GetByUserId(userId string) ([]Token, error)
	GetById(id string) (*Token, error)
	Delete(token *Token) error
}
