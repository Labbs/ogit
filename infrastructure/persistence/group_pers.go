package persistence

import "gorm.io/gorm"

type groupPers struct {
	db *gorm.DB
}

func NewGroupPers(db *gorm.DB) *groupPers {
	return &groupPers{db: db}
}
