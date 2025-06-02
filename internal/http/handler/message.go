package handler

const oopsErr = "Oops! Something went wrong. Please try again later."

type Response struct {
	Message string      `json:"message,omitempty"` // short message for humans
	Data    interface{} `json:"data,omitempty"`    // actual payload (can be nil)
	Error   string      `json:"error,omitempty"`   // error detail (if any)
}
