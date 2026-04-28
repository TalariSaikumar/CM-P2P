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

// BookingHandler serves booking lifecycle and chat.
type BookingHandler struct {
	Svc *service.BookingService
}

type createBookingReq struct {
	CarID        string `json:"car_id" binding:"required,uuid"`
	CustomerNote string `json:"customer_note"`
}

type patchPriceReq struct {
	FinalBookingPrice string `json:"final_booking_price" binding:"required"`
}

type postMessageReq struct {
	Body string `json:"body" binding:"required"`
}

func (h *BookingHandler) Create(c *gin.Context) {
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
	var req createBookingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Please check your input and try again.")
		return
	}
	carID, err := uuid.Parse(req.CarID)
	if err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid car id.")
		return
	}
	b, err := h.Svc.Create(c.Request.Context(), uid, service.CreateBookingInput{
		CarID:        carID,
		CustomerNote: req.CustomerNote,
	})
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"booking": toBookingPublic(b)})
}

func (h *BookingHandler) Mine(c *gin.Context) {
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
	role, _ := middleware.UserRole(c)
	var rows []models.Booking
	switch models.UserRole(role) {
	case models.RoleCustomer:
		rows, err = h.Svc.ListForCustomer(c.Request.Context(), uid)
	case models.RoleOwner:
		rows, err = h.Svc.ListForOwner(c.Request.Context(), uid)
	default:
		httpx.Abort(c, http.StatusForbidden, "FORBIDDEN", "This action is not available for your account type.")
		return
	}
	if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	out := make([]bookingPublic, 0, len(rows))
	for i := range rows {
		out = append(out, toBookingPublic(&rows[i]))
	}
	c.JSON(http.StatusOK, gin.H{"bookings": out})
}

func (h *BookingHandler) Get(c *gin.Context) {
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
	bid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid booking id.")
		return
	}
	b, err := h.Svc.Get(c.Request.Context(), uid, bid)
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"booking": toBookingPublic(b)})
}

func (h *BookingHandler) PatchPrice(c *gin.Context) {
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
	bid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid booking id.")
		return
	}
	var req patchPriceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Please check your input and try again.")
		return
	}
	b, err := h.Svc.PatchFinalPrice(c.Request.Context(), uid, bid, service.PatchFinalPriceInput{
		FinalBookingPrice: req.FinalBookingPrice,
	})
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"booking": toBookingPublic(b)})
}

func (h *BookingHandler) Confirm(c *gin.Context) {
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
	bid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid booking id.")
		return
	}
	b, err := h.Svc.Confirm(c.Request.Context(), uid, bid)
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"booking": toBookingPublic(b)})
}

func (h *BookingHandler) ListMessages(c *gin.Context) {
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
	bid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid booking id.")
		return
	}
	msgs, err := h.Svc.ListMessages(c.Request.Context(), uid, bid)
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	out := make([]messagePublic, 0, len(msgs))
	for i := range msgs {
		out = append(out, toMessagePublic(&msgs[i]))
	}
	c.JSON(http.StatusOK, gin.H{"messages": out})
}

func (h *BookingHandler) PostMessage(c *gin.Context) {
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
	bid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid booking id.")
		return
	}
	var req postMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Please check your input and try again.")
		return
	}
	m, err := h.Svc.PostMessage(c.Request.Context(), uid, bid, service.PostMessageInput{Body: req.Body})
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": toMessagePublic(m)})
}
