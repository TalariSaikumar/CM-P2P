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
	CarID                      string `json:"car_id" binding:"required,uuid"`
	CustomerNote               string `json:"customer_note"`
	RentalFrom                 string `json:"rental_from" binding:"required"`
	RentalTo                   string `json:"rental_to" binding:"required"`
	PickupPoint                string `json:"pickup_point"`
	DropPoint                  string `json:"drop_point"`
	AcknowledgedDepositTerms   bool   `json:"acknowledged_deposit_terms"`
}

type patchPriceReq struct {
	FinalBookingPrice string `json:"final_booking_price" binding:"required"`
}

type patchTripReq struct {
	RentalFrom  string `json:"rental_from" binding:"required"`
	RentalTo    string `json:"rental_to" binding:"required"`
	PickupPoint string `json:"pickup_point" binding:"required"`
	DropPoint   string `json:"drop_point" binding:"required"`
}

type postMessageReq struct {
	Body string `json:"body" binding:"required"`
}

type cancelBookingReq struct {
	Reason string `json:"reason"`
}

type patchHandoverReq struct {
	Phase       string `json:"phase" binding:"required"`
	OdometerKM  int    `json:"odometer_km" binding:"required"`
	FuelPercent *int   `json:"fuel_percent"`
	Notes       string `json:"notes"`
}

type postReviewReq struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
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
	from, err := service.ParseBookingDateTime(req.RentalFrom)
	if err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid rental_from: use RFC3339 or YYYY-MM-DD.")
		return
	}
	to, err := service.ParseBookingDateTime(req.RentalTo)
	if err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid rental_to: use RFC3339 or YYYY-MM-DD.")
		return
	}
	b, err := h.Svc.Create(c.Request.Context(), uid, service.CreateBookingInput{
		CarID:                     carID,
		CustomerNote:              req.CustomerNote,
		RentalFrom:                from,
		RentalTo:                  to,
		PickupPoint:               req.PickupPoint,
		DropPoint:                 req.DropPoint,
		AcknowledgedDepositTerms:  req.AcknowledgedDepositTerms,
	})
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"booking": toBookingPublic(b, h.Svc)})
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
	page, perPage, offset := parseListPagination(c)
	var rows []models.Booking
	var total int64
	switch models.UserRole(role) {
	case models.RoleCustomer:
		rows, total, err = h.Svc.ListForCustomerPaged(c.Request.Context(), uid, offset, perPage)
	case models.RoleOwner:
		rows, total, err = h.Svc.ListForOwnerPaged(c.Request.Context(), uid, offset, perPage)
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
		out = append(out, toBookingPublic(&rows[i], h.Svc))
	}
	c.JSON(http.StatusOK, gin.H{
		"bookings": out,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
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
	c.JSON(http.StatusOK, gin.H{"booking": toBookingPublic(b, h.Svc)})
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
	c.JSON(http.StatusOK, gin.H{"booking": toBookingPublic(b, h.Svc)})
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
	c.JSON(http.StatusOK, gin.H{"booking": toBookingPublic(b, h.Svc)})
}

func (h *BookingHandler) AcceptQuotedPrice(c *gin.Context) {
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
	b, err := h.Svc.CustomerAcceptQuotedPrice(c.Request.Context(), uid, bid)
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"booking": toBookingPublic(b, h.Svc)})
}

func (h *BookingHandler) PatchTrip(c *gin.Context) {
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
	var req patchTripReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Please check your input and try again.")
		return
	}
	from, err := service.ParseBookingDateTime(req.RentalFrom)
	if err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid rental_from: use RFC3339 or YYYY-MM-DD.")
		return
	}
	to, err := service.ParseBookingDateTime(req.RentalTo)
	if err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid rental_to: use RFC3339 or YYYY-MM-DD.")
		return
	}
	b, err := h.Svc.UpdateTripDetails(c.Request.Context(), uid, bid, service.UpdateTripDetailsInput{
		RentalFrom:  from,
		RentalTo:    to,
		PickupPoint: req.PickupPoint,
		DropPoint:   req.DropPoint,
	})
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"booking": toBookingPublic(b, h.Svc)})
}

func (h *BookingHandler) Withdraw(c *gin.Context) {
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
	b, err := h.Svc.Withdraw(c.Request.Context(), uid, bid)
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"booking": toBookingPublic(b, h.Svc)})
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

func (h *BookingHandler) Cancel(c *gin.Context) {
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
	var req cancelBookingReq
	_ = c.ShouldBindJSON(&req)
	b, err := h.Svc.CancelBooking(c.Request.Context(), uid, bid, req.Reason)
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"booking": toBookingPublic(b, h.Svc)})
}

func (h *BookingHandler) PatchHandover(c *gin.Context) {
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
	var req patchHandoverReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Please check your input and try again.")
		return
	}
	b, err := h.Svc.PatchHandover(c.Request.Context(), uid, bid, service.PatchHandoverInput{
		Phase:       req.Phase,
		OdometerKM:  req.OdometerKM,
		FuelPercent: req.FuelPercent,
		Notes:       req.Notes,
	})
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"booking": toBookingPublic(b, h.Svc)})
}

func (h *BookingHandler) UploadHandoverPhoto(c *gin.Context) {
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
	stepStr := c.PostForm("step")
	if stepStr == "" {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Field \"step\" is required.")
		return
	}
	step, err := service.ParseHandoverPhotoStep(stepStr)
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
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
	p, err := h.Svc.UploadHandoverPhoto(c.Request.Context(), uid, bid, step, fh.Filename, ct, src)
	if httpx.AbortService(c, err) {
		return
	}
	if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	b, err := h.Svc.Repo.GetBookingByID(c.Request.Context(), bid)
	if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"photo": handoverPhotoPublic{
			ID:        p.ID.String(),
			Step:      string(p.Step),
			BlobURL:   p.BlobURL,
			CreatedAt: p.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		},
		"booking": toBookingPublic(b, h.Svc),
	})
}

func (h *BookingHandler) PostReview(c *gin.Context) {
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
	var req postReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Please check your input and try again.")
		return
	}
	if _, err := h.Svc.SubmitReview(c.Request.Context(), uid, bid, req.Rating, req.Comment); err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
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
	c.JSON(http.StatusCreated, gin.H{"booking": toBookingPublic(b, h.Svc)})
}

type payBookingReq struct {
	PaymentMethod       string `json:"payment_method" binding:"required"`
	RazorpayOrderID     string `json:"razorpay_order_id"`
	RazorpayPaymentID   string `json:"razorpay_payment_id"`
	RazorpaySignature   string `json:"razorpay_signature"`
}

func (h *BookingHandler) PaymentPreview(c *gin.Context) {
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
	b, bd, err := h.Svc.CustomerPaymentPreview(c.Request.Context(), uid, bid)
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	sv := h.Svc.SettlementView(b, bd)
	c.JSON(http.StatusOK, gin.H{
		"breakdown": bookingPaymentFromBreakdown(b, bd, sv, h.Svc.PaymentCheckoutInfo()),
		"checkout":  h.Svc.PaymentCheckoutInfo(),
	})
}

func (h *BookingHandler) CreatePaymentOrder(c *gin.Context) {
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
	order, err := h.Svc.CustomerCreateRazorpayOrder(c.Request.Context(), uid, bid)
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"order_id":     order.OrderID,
		"amount_paise": order.AmountPaise,
		"currency":     order.Currency,
		"key_id":       order.KeyID,
	})
}

type putPostTripChargesReq service.OwnerPutPostTripChargesInput

// PutPostTripCharges lets the owner declare post-trip charges (tolls, damage, fines) after return handover.
func (h *BookingHandler) PutPostTripCharges(c *gin.Context) {
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
	var req putPostTripChargesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Please check your input and try again.")
		return
	}
	lines, err := service.ParsePostTripChargeLines(service.OwnerPutPostTripChargesInput(req))
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	b, err := h.Svc.OwnerPutPostTripCharges(c.Request.Context(), uid, bid, lines)
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"booking": toBookingPublic(b, h.Svc)})
}

func (h *BookingHandler) Pay(c *gin.Context) {
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
	var req payBookingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Please check your input and try again.")
		return
	}
	var proof *service.RazorpayPaymentProof
	if req.RazorpayOrderID != "" || req.RazorpayPaymentID != "" || req.RazorpaySignature != "" {
		proof = &service.RazorpayPaymentProof{
			OrderID:   req.RazorpayOrderID,
			PaymentID: req.RazorpayPaymentID,
			Signature: req.RazorpaySignature,
		}
	}
	b, err := h.Svc.CustomerRecordPayment(c.Request.Context(), uid, bid, req.PaymentMethod, proof)
	if err != nil {
		if httpx.AbortService(c, err) {
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"booking": toBookingPublic(b, h.Svc)})
}
