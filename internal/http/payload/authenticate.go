package payload

import (
	"fethcher/internal/core"
	"fmt"

	"github.com/jellydator/validation"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a AuthRequest) Validate() error {
	err := validation.ValidateStruct(&a,
		validation.Field(&a.Username, validation.Required),
		validation.Field(&a.Password, validation.Required),
	)
	if err != nil {
		return fmt.Errorf("validate struct: %w", err)
	}

	return nil
}

func (a AuthRequest) ToMessage() core.AuthMessage {
	return core.AuthMessage{
		Username: a.Username,
		Password: a.Password,
	}
}
