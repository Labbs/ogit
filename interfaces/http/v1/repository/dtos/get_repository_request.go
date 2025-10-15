package dtos

type GetRepositoryRequest struct {
	Identifier string `path:"identifier" validate:"required,uuid4|slug"`
}

type GetRepositoryResponse struct {
	Id          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
}
