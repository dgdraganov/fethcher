package payload

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func DecodePayload(r *http.Request, object any) error {
	var err error

	decoder := json.NewDecoder(r.Body)
	defer func() {
		errClose := r.Body.Close()
		if err == nil {
			err = errClose
		}
	}()

	decoder.DisallowUnknownFields()

	err = decoder.Decode(object)
	if err != nil {
		return fmt.Errorf("decoding json payload: %w", err)
	}

	return nil
}
