package user

type UserGetRequest struct {
	ID string `param:"id"`
}

type UserCreateRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}
type UserUpdateRequest struct {
	ID    uint32 `param:"id" validate:"required"`
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type UserDeleteRequest struct {
	ID uint32 `param:"id" validate:"required"`
}
