package middleware

import (
	"net/http"

	"carmanage/backend/internal/httpx"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// KYCVerified blocks the handler when the authenticated user is not KYC verified.
// Expects "user_id" (uuid string) set by auth middleware (to be wired when JWT auth is added).
func KYCVerified(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr, ok := UserID(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httpx.ErrorResponse{
				Errors: []httpx.APIError{{
					Code:    "UNAUTHORIZED",
					Message: "You need to sign in to continue.",
				}},
			})
			return
		}
		uid, err := uuid.Parse(idStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httpx.ErrorResponse{
				Errors: []httpx.APIError{{
					Code:    "UNAUTHORIZED",
					Message: "Your session is invalid. Please sign in again.",
				}},
			})
			return
		}

		var verified bool
		if err := db.Table("users").Select("is_kyc_verified").Where("id = ?", uid).Scan(&verified).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, httpx.ErrorResponse{
				Errors: []httpx.APIError{{
					Code:    "INTERNAL_ERROR",
					Message: "We could not verify your account status. Please try again later.",
				}},
			})
			return
		}

		if !verified {
			c.AbortWithStatusJSON(http.StatusForbidden, httpx.ErrorResponse{
				Errors: []httpx.APIError{{
					Code:    "KYC_REQUIRED",
					Message: "Complete identity verification before using this feature.",
				}},
			})
			return
		}

		c.Next()
	}
}
