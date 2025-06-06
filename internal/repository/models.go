package repository

type Transaction struct {
	TransactionHash   string  `gorm:"size:66;uniqueIndex;not null"`
	TransactionStatus uint64  `gorm:"not null"`
	BlockHash         string  `gorm:"size:66;not null"`
	BlockNumber       uint64  `gorm:"not null;index"`
	From              string  `gorm:"size:42;not null"`
	To                *string `gorm:"size:42"`
	ContractAddress   *string `gorm:"size:42"`
	LogsCount         int     `gorm:"not null;default:0"`
	Input             string  `gorm:"type:text;not null"`
	Value             string  `gorm:"size:100;not null"`
}

type User struct {
	ID           string `gorm:"primaryKey;autoIncrement:false"`
	Username     string `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
}

type UserTransaction struct {
	UserID          string `gorm:"uniqueIndex:idx_user_tx;not null"`
	TransactionHash string `gorm:"uniqueIndex:idx_user_tx;not null"`
}
