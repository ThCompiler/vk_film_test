package response

import "vk_film/internal/pkg/types"

type User struct {
	ID    types.Id `json:"id" swaggertype:"integer" format:"uint64" example:"5"`
	Login string   `json:"login" swaggertype:"string" example:"login"`
	Role  string   `json:"role" swaggertype:"string" example:"user" enums:"user,admin"`
}
