package domain

import "time"

type Session struct {
	Id     string
	UserId string

	User      User `gorm:"foreignKey:UserId;references:Id"`
	UserAgent string
	IpAddress string
	ExpiresAt time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *Session) TableName() string {
	return "session"
}

type SessionPers interface {
	Create(session *Session) error
	GetById(id string) (*Session, error)
}
