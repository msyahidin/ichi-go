package bun

import (
	"context"
	"github.com/uptrace/bun"
	"time"
)

type CoreModel struct {
	ID        int64        `bun:"id,pk,autoincrement"`
	Version   int64        `bun:"versions,notnull,default:0"`
	CreatedAt time.Time    `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt bun.NullTime `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
	DeletedAt bun.NullTime `bun:"deleted_at,soft_delete,nullzero,default:null"`
	CreatedBy int64        `bun:"created_by,notnull,default:0"`
	UpdatedBy int64        `bun:"updated_by,notnull,default:0"`
	DeletedBy int64        `bun:"deleted_by"`
}

type Versioned interface {
	GetVersion() int64
	TouchVersion()
}

type Queryable interface {
	GetDB() *bun.DB
	GetTableName() string
	GetAlias() string
}

var _ bun.BeforeAppendModelHook = (*CoreModel)(nil)
var _ bun.BeforeUpdateHook = (*CoreModel)(nil)

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

func (m *CoreModel) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	// TODO get user from context
	//requestContext := ctx.Value("requestContext")
	//if requestContext == nil {
	//
	//}
	switch query.(type) {
	case *bun.InsertQuery:
		m.CreatedAt = time.Now()
		m.Version = time.Now().UnixNano()
		m.CreatedBy = 0
	case *bun.UpdateQuery:
		m.UpdatedAt = bun.NullTime{Time: time.Now()}
		m.UpdatedBy = 0
	case *bun.DeleteQuery:
		m.DeletedAt = bun.NullTime{Time: time.Now()}
		m.DeletedBy = 0
	default:
		// Do nothing for other query types
	}
	return nil
}

func (m *CoreModel) TouchVersion() {
	m.Version = time.Now().UnixNano()
	m.UpdatedAt = bun.NullTime{Time: time.Now()}
}

func (m *CoreModel) GetVersion() int64 {
	return m.Version
}

func (m *CoreModel) BeforeUpdate(ctx context.Context, query *bun.UpdateQuery) error {
	data := query.GetModel().Value()
	if data == nil {
		return nil
	}
	if core, ok := data.(Versioned); ok {
		query.Where("versions = ?", core.GetVersion())
		core.TouchVersion()
	}

	return nil
}
