package payload

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jellydator/validation"
)

type DecodeValidator struct{}

func (dv DecodeValidator) DecodeAndValidateJSONPayload(r *http.Request, object any) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	decoder.DisallowUnknownFields()
	err := decoder.Decode(object)
	if err != nil {
		return fmt.Errorf("decoding json payload: %w", err)
	}
	return dv.validatePayload(object)
}

func (dv *DecodeValidator) validatePayload(object any) error {
	t, ok := object.(validation.Validatable)
	if !ok {
		// nothing to validate
		return nil
	}

	if err := t.Validate(); err != nil {
		return fmt.Errorf("validating payload: %w", err)
	}

	return nil
}
