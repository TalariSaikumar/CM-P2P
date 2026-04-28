package httpx

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Abort writes a structured error response and stops the chain.
func Abort(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, ErrorResponse{
		Errors: []APIError{{Code: code, Message: message}},
	})
}

// AbortService writes from a ServiceError.
func AbortService(c *gin.Context, err error) bool {
	var se *ServiceError
	if errors.As(err, &se) {
		Abort(c, se.Status, se.Code, se.Message)
		return true
	}
	return false
}

// AbortUnexpected maps unknown errors to a safe 500 response.
func AbortUnexpected(c *gin.Context, err error) {
	Abort(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong on our side. Please try again in a moment.")
}
