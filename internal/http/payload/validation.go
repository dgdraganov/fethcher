package payload

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Decoder struct{}

func (dv Decoder) DecodeJSONPayload(r *http.Request, object any) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	decoder.DisallowUnknownFields()
	err := decoder.Decode(object)
	if err != nil {
		return fmt.Errorf("decoding json payload: %w", err)
	}
	return nil
}
