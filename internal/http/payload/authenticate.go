package payload

import (
	"fethcher/internal/core"

	"github.com/jellydator/validation"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a AuthRequest) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.Username, validation.Required),
		validation.Field(&a.Password, validation.Required),
	)
}

func (a AuthRequest) ToCoreAuthMessage() core.AuthMessage {
	return core.AuthMessage{
		Username: a.Username,
		Password: a.Password,
	}
}
