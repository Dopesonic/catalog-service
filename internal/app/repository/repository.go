package repository

import "context"

type Migrate interface {
	Migrate(ctx context.Context) (oldVer, newVer int64, err error)
}
