package user

import (
	"ichi-go/internal/infra/database/bun"

	upbun "github.com/uptrace/bun"
)

type UserModel struct {
	upbun.BaseModel `bun:"table:users,alias:u" dto:"ignore"`
	bun.CoreModel   `bun:",embed"`
	Name            string `bun:"name,unique,notnull"`
	Email           string `bun:"email,unique,notnull"`
	Password        string `bun:"password,notnull" json:"-"`
}
