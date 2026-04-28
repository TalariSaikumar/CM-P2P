package handler

import (
	"net/http"

	"carmanage/backend/internal/httpx"
	"carmanage/backend/internal/middleware"
	"carmanage/backend/internal/models"
	"carmanage/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler serves profile and KYC upload routes.
type UserHandler struct {
	Svc *service.UserService
}

type updateMeReq struct {
	FullName             *string `json:"full_name"`
	AadhaarNumber        *string `json:"aadhaar_number"`
	PhoneNumber          *string `json:"phone_number"`
	Address              *string `json:"address"`
	DrivingLicenseNumber *string `json:"driving_license_number"`
}

func (h *UserHandler) Me(c *gin.Context) {
	idStr, ok := middleware.UserID(c)
	if !ok {
		httpx.Abort(c, http.StatusUnauthorized, "UNAUTHORIZED", "You need to sign in to continue.")
		return
	}
	uid, err := uuid.Parse(idStr)
	if err != nil {
		httpx.Abort(c, http.StatusUnauthorized, "UNAUTHORIZED", "Your session is invalid. Please sign in again.")
		return
	}
	u, err := h.Svc.GetByID(c.Request.Context(), uid)
	if httpx.AbortService(c, err) {
		return
	}
	if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": toUserPublic(u)})
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	idStr, ok := middleware.UserID(c)
	if !ok {
		httpx.Abort(c, http.StatusUnauthorized, "UNAUTHORIZED", "You need to sign in to continue.")
		return
	}
	uid, err := uuid.Parse(idStr)
	if err != nil {
		httpx.Abort(c, http.StatusUnauthorized, "UNAUTHORIZED", "Your session is invalid. Please sign in again.")
		return
	}
	var req updateMeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Please check your input and try again.")
		return
	}
	u, err := h.Svc.UpdateProfile(c.Request.Context(), uid, service.UpdateProfileInput{
		FullName:             req.FullName,
		AadhaarNumber:        req.AadhaarNumber,
		PhoneNumber:          req.PhoneNumber,
		Address:              req.Address,
		DrivingLicenseNumber: req.DrivingLicenseNumber,
	})
	if httpx.AbortService(c, err) {
		return
	}
	if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": toUserPublic(u)})
}

func (h *UserHandler) CompleteKYC(c *gin.Context) {
	idStr, ok := middleware.UserID(c)
	if !ok {
		httpx.Abort(c, http.StatusUnauthorized, "UNAUTHORIZED", "You need to sign in to continue.")
		return
	}
	uid, err := uuid.Parse(idStr)
	if err != nil {
		httpx.Abort(c, http.StatusUnauthorized, "UNAUTHORIZED", "Your session is invalid. Please sign in again.")
		return
	}
	u, err := h.Svc.CompleteKYC(c.Request.Context(), uid)
	if httpx.AbortService(c, err) {
		return
	}
	if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": toUserPublic(u)})
}

func (h *UserHandler) UploadKYC(c *gin.Context) {
	idStr, ok := middleware.UserID(c)
	if !ok {
		httpx.Abort(c, http.StatusUnauthorized, "UNAUTHORIZED", "You need to sign in to continue.")
		return
	}
	uid, err := uuid.Parse(idStr)
	if err != nil {
		httpx.Abort(c, http.StatusUnauthorized, "UNAUTHORIZED", "Your session is invalid. Please sign in again.")
		return
	}
	kindStr := c.PostForm("kind")
	if kindStr == "" {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Field \"kind\" is required (e.g. AADHAAR_FRONT).")
		return
	}
	kind := models.KYCAttachmentKind(kindStr)
	switch kind {
	case models.KYCKindAadhaarFront, models.KYCKindAadhaarBack, models.KYCKindDrivingLicense, models.KYCKindAddressProof, models.KYCKindOther:
	default:
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Unknown document kind.")
		return
	}
	fh, err := c.FormFile("file")
	if err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "A file field named \"file\" is required.")
		return
	}
	src, err := fh.Open()
	if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	defer src.Close()

	ct := fh.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}

	a, err := h.Svc.UploadKYCAttachment(c.Request.Context(), uid, kind, fh.Filename, ct, src)
	if httpx.AbortService(c, err) {
		return
	}
	if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"attachment": kycAttachmentPublic{
		ID:        a.ID.String(),
		Kind:      string(a.Kind),
		BlobURL:   a.BlobURL,
		CreatedAt: a.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}})
}
