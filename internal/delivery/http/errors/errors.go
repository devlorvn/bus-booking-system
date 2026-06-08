package errors

type AppError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}
