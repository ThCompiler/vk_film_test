package user

import "vk_film/internal/pkg/types"

type User struct {
	ID       types.Id
	Login    string
	Password string
	Role     types.Roles
}

type LoginUser struct {
	ID       types.Id
	Password string
}
