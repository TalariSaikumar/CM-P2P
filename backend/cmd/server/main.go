package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	"carmanage/backend/internal/config"
	"carmanage/backend/internal/database"
	"carmanage/backend/internal/handler"
	"carmanage/backend/internal/middleware"
	"carmanage/backend/internal/notify"
	"carmanage/backend/internal/repository"
	"carmanage/backend/internal/service"
	"carmanage/backend/pkg/azureblob"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/logger"
)

func main() {
	_, file, _, _ := runtime.Caller(0)
	backendRoot := filepath.Join(filepath.Dir(file), "..", "..")
	envPath := filepath.Join(backendRoot, ".env")

	cfg, err := config.Load(envPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.NewPostgres(cfg.DatabaseURL, logger.Warn)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	repo := repository.New(db)

	var blob *azureblob.Client
	if cfg.AzureStorageAccount != "" && cfg.AzureStorageKey != "" && cfg.AzureStorageContainer != "" {
		blob, err = azureblob.NewClient(cfg.AzureStorageAccount, cfg.AzureStorageKey, cfg.AzureStorageContainer)
		if err != nil {
			log.Printf("warning: azure blob disabled: %v", err)
			blob = nil
		}
	}

	var sms service.SMSSender
	if cfg.TwilioAccountSID != "" && cfg.TwilioAuthToken != "" && cfg.TwilioFromNumber != "" {
		sms = &notify.TwilioSMS{
			AccountSID: cfg.TwilioAccountSID,
			AuthToken:  cfg.TwilioAuthToken,
			FromNumber: cfg.TwilioFromNumber,
		}
	}

	svcDeps := service.Deps{
		Config: cfg,
		Repo:   repo,
		Blob:   blob,
		SMS:    sms,
	}

	authSvc := &service.AuthService{Deps: svcDeps}
	userSvc := &service.UserService{Deps: svcDeps}
	carSvc := &service.CarService{Deps: svcDeps}
	bookingSvc := &service.BookingService{Deps: svcDeps}

	if gin.Mode() == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())
	r.Use(middleware.GlobalErrorHandler())
	r.MaxMultipartMemory = 32 << 20 // 32 MiB

	api := r.Group("/api")
	handler.RegisterWithDeps(api, handler.Deps{
		Config:  cfg,
		DB:      db,
		Repo:    repo,
		Auth:    authSvc,
		User:    userSvc,
		Car:     carSvc,
		Booking: bookingSvc,
	})

	addr := fmt.Sprintf(":%s", cfg.HTTPPort)
	log.Printf("listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
