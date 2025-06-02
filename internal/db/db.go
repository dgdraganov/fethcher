package db

import (
	"errors"
	"fmt"
	"reflect"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var ErrNotFound = errors.New("record not found")

type PostgresDB struct {
	db *gorm.DB
}

func NewPostgresDB(dsn string) (*PostgresDB, error) {
	var err error
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return &PostgresDB{}, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &PostgresDB{
		db: db,
	}, nil
}

func (f *PostgresDB) MigrateTable(tbl ...any) error {
	err := f.db.AutoMigrate(tbl...)
	if err != nil {
		return fmt.Errorf("failed to migrate table: %w", err)
	}

	return nil
}

func (f *PostgresDB) SaveToTable(records any) error {

	v := reflect.ValueOf(records)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("records type must be pointer to a slice: %T", records)
	}

	slice := v.Elem()
	if slice.Len() == 0 {
		return nil
	}

	var count int64

	elemType := slice.Index(0).Interface()
	if err := f.db.Model(elemType).Count(&count).Error; err != nil {
		return fmt.Errorf("get model count: %w", err)
	}

	if count > 0 {
		return nil
	}

	if err := f.db.Create(records).Error; err != nil {
		return fmt.Errorf("insert to table: %w", err)
	}

	return nil
}

func (f *PostgresDB) CreateDB(dbName string) error {
	if dbName == "" {
		return errors.New("database name cannot be empty")
	}

	sqlDB, err := f.db.DB()
	if err != nil {
		return fmt.Errorf("get sql db conn: %w", err)
	}

	query := `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`

	var exists bool
	if err := sqlDB.QueryRow(query, dbName).Scan(&exists); err != nil {
		return fmt.Errorf("check db exists: %w", err)
	}

	if exists {
		return nil
	}

	// safe to use placeholder since dbName is controlled
	_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		return fmt.Errorf("reate database: %w", err)
	}

	return nil
}

func (f *PostgresDB) GetOneBy(column string, value any, entity any) error {
	query := fmt.Sprintf("%s = ?", column)
	err := f.db.Where(query, value).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("getting record by %q: %w", column, err)
	}
	return nil
}

func (f *PostgresDB) GetAllBy(column string, value any, entity any) error {
	tx := f.db.Where(fmt.Sprintf("%s IN ?", column), value).Find(entity)
	if tx.Error != nil {
		return fmt.Errorf("getting records by %q: %w", column, tx.Error)
	}
	return nil
}
