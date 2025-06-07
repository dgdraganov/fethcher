package repository

import "context"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fake -fake-name Storage . Storage
type Storage interface {
	MigrateTable(tbl ...any) error
	SaveToTable(rctx context.Context, records any) error
	GetOneBy(ctx context.Context, column string, value any, entity any) error
	GetAllBy(ctx context.Context, column string, value any, entity any) error
	GetAll(ctx context.Context, entity any) error
}
