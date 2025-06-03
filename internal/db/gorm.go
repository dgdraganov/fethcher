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

type GormDB struct {
	db *gorm.DB
}

func NewGormDB(dsn string) (*GormDB, error) {
	var err error
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return &GormDB{}, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &GormDB{
		db: db,
	}, nil
}

func (f *GormDB) MigrateModels(models ...any) error {
	err := f.db.AutoMigrate(models...)
	if err != nil {
		return fmt.Errorf("failed to migrate table: %w", err)
	}

	return nil
}

func (f *GormDB) Seed(records any) error {
	v := reflect.ValueOf(records)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("records type must be pointer to a slice: %T", records)
	}

	slice := v.Elem()
	if slice.Len() == 0 {
		return nil
	}

	elemType := slice.Index(0).Interface()
	var count int64
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

// func (f *GormDB) Create(dbName string) error {
// 	if dbName == "" {
// 		return errors.New("database name cannot be empty")
// 	}

// 	sqlDB, err := f.db.DB()
// 	if err != nil {
// 		return fmt.Errorf("get sql db conn: %w", err)
// 	}

// 	var exists bool
// 	query := `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`
// 	if err := sqlDB.QueryRow(query, dbName).Scan(&exists); err != nil {
// 		return fmt.Errorf("check db exists: %w", err)
// 	}

// 	if exists {
// 		return nil
// 	}

// 	// safe to use placeholder since dbName is controlled
// 	_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
// 	if err != nil {
// 		return fmt.Errorf("reate database: %w", err)
// 	}

// 	return nil
// }

func (f *GormDB) GetBy(column string, value any, entity any) error {
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
