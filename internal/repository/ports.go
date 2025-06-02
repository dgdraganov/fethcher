package repository

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fake -fake-name Database . Database
type Database interface {
	MigrateTable(tbl ...any) error
	SeedDB(records any) error
	GetBy(column string, value any, entity any) error
}
