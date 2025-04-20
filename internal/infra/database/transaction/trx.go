package transaction

import (
	"context"
	"ichi-go/internal/infra/database/ent"
)

type Trx interface {
	WithTx(ctx context.Context, fn func(tx *ent.Tx) error) error
}
