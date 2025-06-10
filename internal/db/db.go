package db

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var ErrNotFound = errors.New("record not found")

// PostgresDB is a struct that provides methods to interact with a PostgreSQL database using GORM.
type PostgresDB struct {
	DB *gorm.DB
}

// NewPostgresDB is a constructor function that initializes a new PostgresDB instance.
func NewPostgresDB(dsn string) (*PostgresDB, error) {
	db, err := connectWithRetry(dsn, 10)
	if err != nil {
		return nil, fmt.Errorf("connect to database with retries: %w", err)
	}

	return &PostgresDB{
		DB: db,
	}, nil
}

// MigrateTable migrates the provided tables to the database schema.
func (f *PostgresDB) MigrateTable(tbl ...any) error {
	err := f.DB.AutoMigrate(tbl...)
	if err != nil {
		return fmt.Errorf("failed to migrate table: %w", err)
	}

	return nil
}

// InsertToTable inserts the provided records into the specified table.
func (f *PostgresDB) InsertToTable(ctx context.Context, records any) error {
	if err := f.DB.Create(records).Error; err != nil {
		return fmt.Errorf("insert to table: %w", err)
	}
	return nil
}

// GetOneBy retrieves a single record from the specified table where the given column matches the provided value.
func (f *PostgresDB) GetOneBy(ctx context.Context, column string, value any, entity any) error {
	query := fmt.Sprintf("%s = ?", column)
	err := f.DB.Where(query, value).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("getting record by %q: %w", column, err)
	}
	return nil
}

// GetAllBy retrieves all records from the specified table where the given column matches the provided value.
func (f *PostgresDB) GetAllBy(ctx context.Context, column string, value any, entity any) error {
	tx := f.DB.Where(fmt.Sprintf("%s IN (?)", column), value).Find(entity)
	if tx.Error != nil {
		return fmt.Errorf("getting records by %q: %w", column, tx.Error)
	}
	return nil
}

// GetAll retrieves all records from the specified table and stores them in the provided entity object
func (f *PostgresDB) GetAll(ctx context.Context, entity any) error {
	tx := f.DB.Find(entity)
	if tx.Error != nil {
		return fmt.Errorf("getting all records: %w", tx.Error)
	}
	return nil
}

// SeedTable checks if the table is empty and seeds it with the provided records if it is.
func (f *PostgresDB) SeedTable(ctx context.Context, records any) error {

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
	if err := f.DB.Model(elemType).Count(&count).Error; err != nil {
		return fmt.Errorf("get model count: %w", err)
	}

	if count > 0 {
		return nil
	}

	if err := f.DB.Create(records).Error; err != nil {
		return fmt.Errorf("insert to table: %w", err)
	}

	return nil
}

func connectWithRetry(dsn string, maxRetries int) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			return db, nil
		}
		<-time.After(time.Second * time.Duration(i+1))
	}

	return nil, fmt.Errorf("connect to database after %d retries: %w", maxRetries, err)
}
