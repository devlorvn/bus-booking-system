package errors

import "net/http"

func BadRequest(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Message:    message,
	}
}

func NotFound(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusNotFound,
		Message:    message,
	}
}

func InternalServerError(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Message:    message,
	}
}

func Conflict(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusConflict,
		Message:    message,
	}
}
