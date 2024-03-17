package request

import (
	"github.com/miladibra10/vjson"
	"vk_film/internal/pkg/evjson"
	"vk_film/internal/pkg/time"
)

type CreateActor struct {
	Name     string             `json:"name" swaggertype:"string" example:"Тимоти Шаламе"`
	Sex      string             `json:"sex" swaggertype:"string" example:"male" enums:"male,female"`
	Birthday time.FormattedTime `json:"birthday" swaggertype:"string" format:"date" example:"12.02.2002"`
}

func ValidateCreateActor(data []byte) error {
	schema := evjson.NewSchema(
		vjson.String("name").Required(),
		vjson.String("sex").Choices("male", "female").Required(),
		vjson.String("birthday").Required(),
	)
	return schema.ValidateBytes(data)
}

type UpdateActor struct {
	Name     *string             `json:"name,omitempty" swaggertype:"string" example:"Тимоти Шаламе"`
	Sex      *string             `json:"sex,omitempty" swaggertype:"string" example:"male" enums:"male,female"`
	Birthday *time.FormattedTime `json:"birthday,omitempty" swaggertype:"string" format:"date" example:"12.02.2002"`
}

func ValidateUpdateActor(data []byte) error {
	schema := evjson.NewSchema(
		vjson.String("name"),
		vjson.String("sex").Choices("male", "female"),
		vjson.String("birthday"),
	)
	return schema.ValidateBytes(data)
}
