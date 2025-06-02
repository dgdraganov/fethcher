package repository

type Transaction struct {
	TransactionHash   string  `gorm:"size:66;uniqueIndex;not null"` // 0x + 64 hex chars
	TransactionStatus int     `gorm:"not null"`                     // 1 (success) or 0 (failure)
	BlockHash         string  `gorm:"size:66;not null"`             // 0x + 64 hex chars
	BlockNumber       int64   `gorm:"not null;index"`               // Block number
	From              string  `gorm:"size:42;not null"`             // Ethereum address (0x + 40 hex)
	To                *string `gorm:"size:42"`                      // Nullable Ethereum address
	ContractAddress   *string `gorm:"size:42"`                      // Nullable contract address
	LogsCount         int     `gorm:"not null;default:0"`
	Input             string  `gorm:"type:text;not null"` // Hex encoded input data
	Value             string  `gorm:"size:100;not null"`  // Value in wei (string to handle large numbers)
}

type User struct {
	ID           string `gorm:"primaryKey;autoIncrement:false"`
	Username     string `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
}

type UserTransaction struct {
}
