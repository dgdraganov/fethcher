package payload

import (
	"fmt"
	"regexp"

	"github.com/jellydator/validation"
)

type TransactionsRequest struct {
	Transactions []string
}

func (t TransactionsRequest) Validate() error {
	regex, err := regexp.Compile(`^0x[a-f0-9]+`)
	if err != nil {
		return fmt.Errorf("compile regex: %w", err)
	}

	return validation.ValidateStruct(&t,
		validation.Field(&t.Transactions, validation.Required),
		validation.Field(&t.Transactions, validation.Each(validation.Match(regex))),
	)
}
