package handler

import (
	"errors"
	"net/http"

	"carmanage/backend/internal/httpx"
	"carmanage/backend/internal/middleware"
	"carmanage/backend/internal/models"
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

type airbagDetailReq struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
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

	ModelYear       int               `json:"model_year" binding:"required"`
	Color           string            `json:"color" binding:"required"`
	FuelType        string            `json:"fuel_type" binding:"required"`
	Transmission    string            `json:"transmission" binding:"required"`
	MileageKm       int               `json:"mileage_km" binding:"min=0"`
	NumSeats        int               `json:"num_seats" binding:"required,min=1,max=20"`
	Airbags         bool              `json:"airbags"`
	AirbagCount     int               `json:"airbag_count"`
	AirbagDetails   []airbagDetailReq `json:"airbag_details"`
	CameraType      string            `json:"camera_type"`
	AirConditioning bool              `json:"air_conditioning"`
	CruiseControl   bool              `json:"cruise_control"`
	OpenRoof        bool              `json:"open_roof"`
	Navigation      bool              `json:"navigation"`
	Speakers        bool              `json:"speakers"`
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

	ModelYear       *int               `json:"model_year"`
	Color           *string            `json:"color"`
	FuelType        *string            `json:"fuel_type"`
	Transmission    *string            `json:"transmission"`
	MileageKm       *int               `json:"mileage_km"`
	NumSeats        *int               `json:"num_seats"`
	Airbags         *bool              `json:"airbags"`
	AirbagCount     *int               `json:"airbag_count"`
	AirbagDetails   *[]airbagDetailReq `json:"airbag_details"`
	CameraType      *string            `json:"camera_type"`
	AirConditioning *bool              `json:"air_conditioning"`
	CruiseControl   *bool              `json:"cruise_control"`
	OpenRoof        *bool              `json:"open_roof"`
	Navigation      *bool              `json:"navigation"`
	Speakers        *bool              `json:"speakers"`
}

func (h *CarHandler) Search(c *gin.Context) {
	loc := c.Query("location")
	model := c.Query("model")
	page, perPage, offset := parseListPagination(c)
	cars, total, err := h.Repo.SearchCarsPaged(c.Request.Context(), loc, model, offset, perPage)
	if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	out := make([]carPublic, 0, len(cars))
	for i := range cars {
		out = append(out, toCarPublic(&cars[i]))
	}
	c.JSON(http.StatusOK, gin.H{
		"cars":     out,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
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
	abDetails := make([]models.AirbagDetail, 0, len(req.AirbagDetails))
	for _, d := range req.AirbagDetails {
		abDetails = append(abDetails, models.AirbagDetail{Type: d.Type, Count: d.Count})
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

		ModelYear:       req.ModelYear,
		Color:           req.Color,
		FuelType:        req.FuelType,
		Transmission:    req.Transmission,
		MileageKm:       req.MileageKm,
		NumSeats:        req.NumSeats,
		Airbags:         req.Airbags,
		AirbagCount:     req.AirbagCount,
		AirbagDetails:   abDetails,
		CameraType:      req.CameraType,
		AirConditioning: req.AirConditioning,
		CruiseControl:   req.CruiseControl,
		OpenRoof:        req.OpenRoof,
		Navigation:      req.Navigation,
		Speakers:        req.Speakers,
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

func (h *CarHandler) GetForOwner(c *gin.Context) {
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
	car, err := h.Repo.GetCarByID(c.Request.Context(), carID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpx.Abort(c, http.StatusNotFound, "NOT_FOUND", "We could not find what you were looking for.")
			return
		}
		httpx.AbortUnexpected(c, err)
		return
	}
	if car.OwnerID != uid {
		httpx.Abort(c, http.StatusForbidden, "FORBIDDEN", "You do not have access to this resource.")
		return
	}
	booked, err := h.Svc.CarBookedForCurrentUTCDate(c.Request.Context(), carID)
	if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	c.JSON(http.StatusOK, CarMineEntry{carPublic: toCarPublic(car), BookedForCurrentDate: booked})
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
	page, perPage, offset := parseListPagination(c)
	cars, total, err := h.Repo.ListCarsForOwnerPaged(c.Request.Context(), uid, offset, perPage)
	if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	carIDs := make([]uuid.UUID, 0, len(cars))
	for i := range cars {
		carIDs = append(carIDs, cars[i].ID)
	}
	bookedIDs, err := h.Svc.ListCarIDsBookedTodayUTC(c.Request.Context(), carIDs)
	if err != nil {
		httpx.AbortUnexpected(c, err)
		return
	}
	bookedSet := make(map[uuid.UUID]struct{}, len(bookedIDs))
	for _, id := range bookedIDs {
		bookedSet[id] = struct{}{}
	}
	out := make([]CarMineEntry, 0, len(cars))
	for i := range cars {
		_, booked := bookedSet[cars[i].ID]
		out = append(out, CarMineEntry{carPublic: toCarPublic(&cars[i]), BookedForCurrentDate: booked})
	}
	c.JSON(http.StatusOK, gin.H{
		"cars":     out,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
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
	var abPatch *[]models.AirbagDetail
	if req.AirbagDetails != nil {
		lst := make([]models.AirbagDetail, 0, len(*req.AirbagDetails))
		for _, d := range *req.AirbagDetails {
			lst = append(lst, models.AirbagDetail{Type: d.Type, Count: d.Count})
		}
		abPatch = &lst
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

		ModelYear:       req.ModelYear,
		Color:           req.Color,
		FuelType:        req.FuelType,
		Transmission:    req.Transmission,
		MileageKm:       req.MileageKm,
		NumSeats:        req.NumSeats,
		Airbags:         req.Airbags,
		AirbagCount:     req.AirbagCount,
		AirbagDetails:   abPatch,
		CameraType:      req.CameraType,
		AirConditioning: req.AirConditioning,
		CruiseControl:   req.CruiseControl,
		OpenRoof:        req.OpenRoof,
		Navigation:      req.Navigation,
		Speakers:        req.Speakers,
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
