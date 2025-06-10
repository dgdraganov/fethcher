package core

type TransactionRecord struct {
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

type AuthMessage struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
