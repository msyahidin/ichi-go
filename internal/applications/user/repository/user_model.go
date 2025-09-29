package repository

import (
	upbun "github.com/uptrace/bun"
	"ichi-go/internal/infra/database/bun"
)

type UserModel struct {
	bun.CoreModel
	bun.AuditModel
	upbun.BaseModel `bun:"table:users,alias:u"`

	Name     string `bun:"name,unique,notnull"`
	Email    string `bun:"email,unique,notnull"`
	Password string `bun:"password,notnull" json:"-"`
}
