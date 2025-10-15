package domain

import "time"

type Group struct {
	Id          string
	Name        string
	Description string

	Role Role `gorm:"type:role;default:'user'"`

	// Owner is the username of the user who owns the group
	OwnerId string
	// OwnerUser is the user who owns the group
	OwnerUser User `gorm:"foreignKey:OwnerId;references:Id"`

	// Members is the list of users who are members of the group
	Members []User `gorm:"many2many:group_members;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (g *Group) TableName() string {
	return "group"
}

type GroupPers interface{}
