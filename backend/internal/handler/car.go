package handler

import (
	"errors"
	"net/http"

	"carmanage/backend/internal/httpx"
	"carmanage/backend/internal/middleware"
	"carmanage/backend/internal/repository"
	"carmanage/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CarHandler serves listing CRUD and search.
type CarHandler struct {
	Svc  *service.CarService
	Repo *repository.DB
}

type createCarReq struct {
	CarName            string `json:"car_name" binding:"required"`
	CarModel           string `json:"car_model" binding:"required"`
	CarNumber          string `json:"car_number" binding:"required"`
	RegistrationNumber string `json:"registration_number" binding:"required"`
	EngineNumber       string `json:"engine_number" binding:"required"`
	PricePerHour       string `json:"price_per_hour" binding:"required"`
	PricePerDay        string `json:"price_per_day" binding:"required"`
	PricePerKm         string `json:"price_per_km" binding:"required"`
	Location           string `json:"location" binding:"required"`
}

type patchCarReq struct {
	CarName            *string `json:"car_name"`
	CarModel           *string `json:"car_model"`
	CarNumber          *string `json:"car_number"`
	RegistrationNumber *string `json:"registration_number"`
	EngineNumber       *string `json:"engine_number"`
	PricePerHour       *string `json:"price_per_hour"`
	PricePerDay        *string `json:"price_per_day"`
	PricePerKm         *string `json:"price_per_km"`
	Location           *string `json:"location"`
	IsActive           *bool   `json:"is_active"`
}

func (h *CarHandler) Search(c *gin.Context) {
	loc := c.Query("location")
	model := c.Query("model")
	cars, err := h.Repo.SearchCars(c.Request.Context(), loc, model)
	if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	out := make([]carPublic, 0, len(cars))
	for i := range cars {
		out = append(out, toCarPublic(&cars[i]))
	}
	c.JSON(http.StatusOK, gin.H{"cars": out})
}

func (h *CarHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid car id.")
		return
	}
	car, err := h.Repo.GetCarByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpx.Abort(c, http.StatusNotFound, "NOT_FOUND", "We could not find what you were looking for.")
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"car": toCarPublic(car)})
}

func (h *CarHandler) Create(c *gin.Context) {
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
	var req createCarReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Please check your input and try again.")
		return
	}
	car, err := h.Svc.Create(c.Request.Context(), uid, service.CreateCarInput{
		CarName:            req.CarName,
		CarModel:           req.CarModel,
		CarNumber:          req.CarNumber,
		RegistrationNumber: req.RegistrationNumber,
		EngineNumber:       req.EngineNumber,
		PricePerHour:       req.PricePerHour,
		PricePerDay:        req.PricePerDay,
		PricePerKm:         req.PricePerKm,
		Location:           req.Location,
	})
	if httpx.AbortService(c, err) {
		return
	}
	if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"car": toCarPublic(car)})
}

func (h *CarHandler) Mine(c *gin.Context) {
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
	cars, err := h.Repo.ListCarsForOwner(c.Request.Context(), uid)
	if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	out := make([]carPublic, 0, len(cars))
	for i := range cars {
		out = append(out, toCarPublic(&cars[i]))
	}
	c.JSON(http.StatusOK, gin.H{"cars": out})
}

func (h *CarHandler) Update(c *gin.Context) {
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
	carID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid car id.")
		return
	}
	var req patchCarReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Please check your input and try again.")
		return
	}
	car, err := h.Svc.Update(c.Request.Context(), uid, carID, service.UpdateCarInput{
		CarName:            req.CarName,
		CarModel:           req.CarModel,
		CarNumber:          req.CarNumber,
		RegistrationNumber: req.RegistrationNumber,
		EngineNumber:       req.EngineNumber,
		PricePerHour:       req.PricePerHour,
		PricePerDay:        req.PricePerDay,
		PricePerKm:         req.PricePerKm,
		Location:           req.Location,
		IsActive:           req.IsActive,
	})
	if httpx.AbortService(c, err) {
		return
	}
	if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"car": toCarPublic(car)})
}

func (h *CarHandler) Delete(c *gin.Context) {
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
	carID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid car id.")
		return
	}
	if err := h.Svc.Delete(c.Request.Context(), uid, carID); httpx.AbortService(c, err) {
		return
	} else if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *CarHandler) UploadImage(c *gin.Context) {
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
	carID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Abort(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid car id.")
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
	img, err := h.Svc.AddImage(c.Request.Context(), uid, carID, fh.Filename, ct, src)
	if httpx.AbortService(c, err) {
		return
	}
	if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"image": carImagePublic{
		ID:        img.ID.String(),
		BlobURL:   img.BlobURL,
		SortOrder: img.SortOrder,
	}})
}
