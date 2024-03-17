package request

import (
	"github.com/miladibra10/vjson"
	"vk_film/internal/pkg/evjson"
	"vk_film/internal/pkg/time"
	"vk_film/internal/pkg/types"
)

type CreateFilm struct {
	Name        string             `json:"name" swaggertype:"string" example:"Dune"`
	Description string             `json:"description" swaggertype:"string" example:"Futuristic film"`
	DataPublish time.FormattedTime `json:"data_publish" swaggertype:"string" format:"date" example:"12.02.2023"`
	Rating      types.Rating       `json:"rating" swaggertype:"integer" format:"uint8" example:"9"`
	Actors      []types.Id         `json:"actors"`
}

func ValidateCreateFilm(data []byte) error {
	schema := evjson.NewSchema(
		vjson.String("name").MinLength(1).MaxLength(150).Required(),
		vjson.String("description").MaxLength(1000).Required(),
		vjson.String("data_publish").Required(),
		vjson.Integer("rating").Range(0, 10).Required(),
		vjson.Array("actors", vjson.Integer("item").Positive()).Required(),
	)
	return schema.ValidateBytes(data)
}

type UpdateFilm struct {
	Name        *string             `json:"name,omitempty" swaggertype:"string" example:"Dune"`
	Description *string             `json:"description,omitempty" swaggertype:"string" example:"Futuristic film"`
	DataPublish *time.FormattedTime `json:"data_publish,omitempty" swaggertype:"string" format:"date" example:"12.02.2023"`
	Rating      *types.Rating       `json:"rating,omitempty" swaggertype:"integer" format:"uint8" example:"9"`
	Actors      *[]types.Id         `json:"actors,omitempty"`
}

func ValidateUpdateFilm(data []byte) error {
	schema := evjson.NewSchema(
		vjson.String("name").MinLength(1).MaxLength(150),
		vjson.String("description").MaxLength(1000),
		vjson.String("data_publish"),
		vjson.Integer("rating").Range(0, 10),
		vjson.Array("actors", vjson.Integer("item").Positive()),
	)
	return schema.ValidateBytes(data)
}
