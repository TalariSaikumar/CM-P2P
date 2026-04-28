package middleware

import (
	"log"
	"net/http"

	"carmanage/backend/internal/httpx"

	"github.com/gin-gonic/gin"
)

// GlobalErrorHandler returns structured JSON for panics and unhandled errors.
func GlobalErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				c.AbortWithStatusJSON(http.StatusInternalServerError, httpx.ErrorResponse{
					Errors: []httpx.APIError{{
						Code:    "INTERNAL_ERROR",
						Message: "Something went wrong on our side. Please try again in a moment.",
					}},
				})
			}
		}()
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last()
		status := c.Writer.Status()
		if status == 0 || status == http.StatusOK {
			status = http.StatusBadRequest
		}

		msg := err.Error()
		if msg == "" {
			msg = "The request could not be processed. Please check your input and try again."
		}

		c.AbortWithStatusJSON(status, httpx.ErrorResponse{
			Errors: []httpx.APIError{{
				Code:    "REQUEST_ERROR",
				Message: msg,
			}},
		})
	}
}
