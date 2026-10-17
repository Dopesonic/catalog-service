package rcpostgres

import (
	"context"

	"github.com/uptrace/bun"
)

type contextKeyTx struct{}

func getTxFromContext(ctx context.Context) bun.Tx {
	value, _ := ctx.Value(contextKeyTx{}).(bun.Tx)
	return value
}

func setTxToContext(ctx context.Context, tx bun.Tx) context.Context {
	return context.WithValue(ctx, contextKeyTx{}, tx)
}
