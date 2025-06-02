package repository

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fake -fake-name Storage . Storage
type Storage interface {
	MigrateTable(tbl ...any) error
	SaveToTable(records any) error
	GetOneBy(column string, value any, entity any) error
	GetAllBy(column string, value any, entity any) error
}
