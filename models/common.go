package models

type Common struct {
	CreatedAt uint `json:"created_at" form:"created_at"`
	UpdatedAt uint `json:"updated_at" form:"updated_at"`
	DeletedAt *int `json:"deleted_at" form:"deleted_at"`
}