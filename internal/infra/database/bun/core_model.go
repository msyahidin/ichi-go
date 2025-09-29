package bun

import (
	"github.com/uptrace/bun"
	"time"
)

type CoreModel struct {
	ID        int64        `bun:"id,pk,autoincrement"`
	CreatedAt time.Time    `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time    `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
	DeletedAt bun.NullTime `bun:"deleted_at,soft_delete,nullzero"`
}

type AuditModel struct {
	CreatedBy int64 `bun:"created_by,notnull,default:0"`
	UpdatedBy int64 `bun:"updated_by,notnull,default:0"`
	DeletedBy int64 `bun:"deleted_by"`
}

type Queryable interface {
	GetDB() *bun.DB
	GetTableName() string
	GetAlias() string
}

type QueryBuilder struct {
	db    *bun.DB
	query *bun.SelectQuery
}

func NewQueryBuilder(db *bun.DB, model interface{}) *QueryBuilder {
	return &QueryBuilder{
		db:    db,
		query: db.NewSelect().Model(model),
	}
}
