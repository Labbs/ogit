package domain

import (
	"time"
)

type SSHKey struct {
	Id        string
	UserId    string
	RepoId    string
	PublicKey string
	Type      SSHType

	Repository Repository `gorm:"foreignKey:RepoId"`
	User       User       `gorm:"foreignKey:UserId"`

	OwnerType SSHKeyOwnerType

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (SSHKey) TableName() string {
	return "ssh_key"
}

type SSHType string

// ed25519, rsa, ecdsa, dsa
const (
	SSHTypeED25519 SSHType = "ed25519"
	SSHTypeRSA     SSHType = "rsa"
	SSHTypeECDSA   SSHType = "ecdsa"
	SSHTypeDSA     SSHType = "dsa"
)

type SSHKeyOwnerType string

const (
	SSHKeyTypeUser       SSHKeyOwnerType = "user"
	SSHKeyTypeRepository SSHKeyOwnerType = "repository"
)

type SSHKeyPers interface{}
