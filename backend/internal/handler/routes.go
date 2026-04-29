package handler

import (
	"carmanage/backend/internal/config"
	"carmanage/backend/internal/middleware"
	"carmanage/backend/internal/models"
	"carmanage/backend/internal/repository"
	"carmanage/backend/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Deps bundles initialized services for HTTP layer.
type Deps struct {
	Config *config.Config
	DB     *gorm.DB
	Repo   *repository.DB

	Auth    *service.AuthService
	User    *service.UserService
	Car     *service.CarService
	Booking *service.BookingService
}

// RegisterWithDeps attaches routes using pre-built services.
func RegisterWithDeps(r *gin.RouterGroup, d Deps) {
	authH := &AuthHandler{Svc: d.Auth}
	userH := &UserHandler{Svc: d.User}
	carH := &CarHandler{Svc: d.Car, Repo: d.Repo}
	bookH := &BookingHandler{Svc: d.Booking}
	health := &Health{DB: d.DB}

	auth := r.Group("/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
	}

	r.GET("/health", health.Get)

	r.GET("/cars", carH.Search)
	r.GET("/cars/:id", carH.Get)

	au := r.Group("")
	au.Use(middleware.AuthRequired(d.Config))
	{
		au.GET("/me", userH.Me)
		au.PUT("/me", userH.UpdateMe)
		au.POST("/me/complete-kyc", userH.CompleteKYC)
		au.POST("/me/kyc-attachments", userH.UploadKYC)

		ownerCars := au.Group("")
		ownerCars.Use(middleware.RequireRole(models.RoleOwner))
		ownerCars.Use(middleware.KYCVerified(d.DB))
		{
			ownerCars.POST("/cars", carH.Create)
			ownerCars.GET("/cars/mine", carH.Mine)
			ownerCars.PATCH("/cars/:id", carH.Update)
			ownerCars.DELETE("/cars/:id", carH.Delete)
			ownerCars.POST("/cars/:id/images", carH.UploadImage)
		}

		custBook := au.Group("")
		custBook.Use(middleware.RequireRole(models.RoleCustomer))
		custBook.Use(middleware.KYCVerified(d.DB))
		{
			custBook.POST("/bookings", bookH.Create)
			custBook.PATCH("/bookings/:id/trip", bookH.PatchTrip)
			custBook.POST("/bookings/:id/withdraw", bookH.Withdraw)
		}

		au.GET("/bookings/mine", bookH.Mine)
		au.GET("/bookings/:id", bookH.Get)
		au.GET("/bookings/:id/messages", bookH.ListMessages)
		au.POST("/bookings/:id/messages", bookH.PostMessage)

		ownerPrice := au.Group("")
		ownerPrice.Use(middleware.RequireRole(models.RoleOwner))
		ownerPrice.Use(middleware.KYCVerified(d.DB))
		{
			ownerPrice.PATCH("/bookings/:id/price", bookH.PatchPrice)
			ownerPrice.POST("/bookings/:id/confirm", bookH.Confirm)
		}
	}
}
