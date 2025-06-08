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

type PostgresDB struct {
	DB *gorm.DB
}

func NewPostgresDB(dsn string) (*PostgresDB, error) {
	db, err := connectWithRetry(dsn, 10)
	if err != nil {
		return nil, fmt.Errorf("connect to database with retries: %w", err)
	}

	return &PostgresDB{
		DB: db,
	}, nil
}

func (f *PostgresDB) MigrateTable(tbl ...any) error {
	err := f.DB.AutoMigrate(tbl...)
	if err != nil {
		return fmt.Errorf("failed to migrate table: %w", err)
	}

	return nil
}

func (f *PostgresDB) InsertToTable(ctx context.Context, records any) error {
	if err := f.DB.Create(records).Error; err != nil {
		return fmt.Errorf("insert to table: %w", err)
	}
	return nil
}

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

func (f *PostgresDB) GetAllBy(ctx context.Context, column string, value any, entity any) error {
	tx := f.DB.Where(fmt.Sprintf("%s IN (?)", column), value).Find(entity)
	if tx.Error != nil {
		return fmt.Errorf("getting records by %q: %w", column, tx.Error)
	}
	return nil
}

func (f *PostgresDB) GetAll(ctx context.Context, entity any) error {
	tx := f.DB.Find(entity)
	if tx.Error != nil {
		return fmt.Errorf("getting all records: %w", tx.Error)
	}
	return nil
}

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
