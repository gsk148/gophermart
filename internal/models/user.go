package models

type User struct {
	ID       uint
	Login    string `json:"login" validate:"required,min=3"`
	Password string `json:"password" validate:"required"`
}
