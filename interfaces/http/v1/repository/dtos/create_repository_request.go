package dtos

type CreateRepositoryRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
}

type CreateRepositoryResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}
