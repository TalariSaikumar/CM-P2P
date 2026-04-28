package httpx

// APIError is the standard JSON error shape returned by the global error handler.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse wraps one or more API errors for consistent responses.
type ErrorResponse struct {
	Errors []APIError `json:"errors"`
}
