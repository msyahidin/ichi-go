package model

import (
	upbun "github.com/uptrace/bun"
)

type User struct {
	upbun.BaseModel `bun:"table:users,alias:u" dto:"ignore"`
	CoreModel       `bun:",embed"`
	Name            string `bun:"name,unique,notnull"`
	Email           string `bun:"email,unique,notnull"`
	Password        string `bun:"password,notnull" json:"-"`
}
