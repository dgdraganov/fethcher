package ethereum

type TxResult struct {
	Transaction *Transaction
	Error       error
}

type Transaction struct {
	TransactionHash   string
	TransactionStatus uint64
	BlockHash         string
	BlockNumber       uint64
	From              string
	To                *string
	ContractAddress   *string
	LogsCount         int
	Input             string
	Value             string
}
