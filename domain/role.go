package domain

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
	RoleGest  Role = "guest"
)
