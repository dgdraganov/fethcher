package models

type User struct {
	ID           string `gorm:"primaryKey;autoIncrement:false"`
	Username     string `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
}
