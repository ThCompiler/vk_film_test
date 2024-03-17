package request

import (
	"github.com/miladibra10/vjson"
	"vk_film/internal/pkg/evjson"
)

type CreateUser struct {
	Login    string `json:"login" swaggertype:"string" example:"login"`
	Password string `json:"password" swaggertype:"string" example:"password"`
	Role     string `json:"role,omitempty" swaggertype:"string" example:"user" enums:"user,admin" default:"user"`
}

func ValidateCreateUser(data []byte) error {
	schema := evjson.NewSchema(
		vjson.String("login").Required(),
		vjson.String("password").Required(),
		vjson.String("role").Choices("user", "admin"),
	)

	return schema.ValidateBytes(data)
}

type Login struct {
	Login    string `json:"login" swaggertype:"string" example:"login"`
	Password string `json:"password" swaggertype:"string" example:"password"`
}

func ValidateLogin(data []byte) error {
	schema := evjson.NewSchema(
		vjson.String("login").Required(),
		vjson.String("password").Required(),
	)

	return schema.ValidateBytes(data)
}

type UpdateRole struct {
	Role string `json:"role" swaggertype:"string" example:"user" enums:"user,admin"`
}

func ValidateUpdateRole(data []byte) error {
	schema := evjson.NewSchema(
		vjson.String("role").Choices("user", "admin").Required(),
	)

	return schema.ValidateBytes(data)
}
