package storage

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fake -fake-name Database . Database
type Database interface {
	MigrateModels(models ...any) error
	Seed(records any) error
	GetBy(column string, value any, entity any) error
}
