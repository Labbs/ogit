package domain

import "time"

type Config struct {
	Id        string
	Key       string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
