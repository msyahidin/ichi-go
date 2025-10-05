package repository

import (
	upbun "github.com/uptrace/bun"
	"ichi-go/internal/infra/database/bun"
)

type UserModel struct {
	upbun.BaseModel `bun:"table:users,alias:u"`
	bun.CoreModel   `bun:"embed"`
	Name            string `bun:"name,unique,notnull"`
	Email           string `bun:"email,unique,notnull"`
	Password        string
}
