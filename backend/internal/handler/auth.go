package handler

import (
	"errors"
	"net/http"

	"carmanage/backend/internal/httpx"
	"carmanage/backend/internal/models"
	"carmanage/backend/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler exposes register/login.
type AuthHandler struct {
	Svc *service.AuthService
}

type registerReq struct {
	Email                string  `json:"email" binding:"required,email"`
	Password             string  `json:"password" binding:"required,min=8"`
	Role                 string  `json:"role" binding:"required,oneof=CUSTOMER OWNER"`
	FullName             string  `json:"full_name" binding:"required"`
	AadhaarNumber        string  `json:"aadhaar_number" binding:"required"`
	PhoneNumber          string  `json:"phone_number" binding:"required"`
	Address              string  `json:"address" binding:"required"`
	DrivingLicenseNumber *string `json:"driving_license_number"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Please check your input and try again.")
		return
	}
	u, token, err := h.Svc.Register(c.Request.Context(), service.RegisterInput{
		Email:                req.Email,
		Password:             req.Password,
		Role:                 models.UserRole(req.Role),
		FullName:             req.FullName,
		AadhaarNumber:        req.AadhaarNumber,
		PhoneNumber:          req.PhoneNumber,
		Address:              req.Address,
		DrivingLicenseNumber: req.DrivingLicenseNumber,
	})
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user":  toUserPublic(u),
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Please check your input and try again.")
		return
	}
	u, token, err := h.Svc.Login(c.Request.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		var se *httpx.ServiceError
		if errors.As(err, &se) {
			httpx.AbortService(c, err)
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  toUserPublic(u),
	})
}
