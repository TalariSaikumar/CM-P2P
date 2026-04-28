package service

import (
	"context"

	"carmanage/backend/internal/config"
	"carmanage/backend/internal/repository"
	"carmanage/backend/pkg/azureblob"
)

// SMSSender abstracts outbound SMS (Twilio in production).
type SMSSender interface {
	Send(ctx context.Context, toE164, body string) error
}

// Deps bundles cross-cutting dependencies for services.
type Deps struct {
	Config *config.Config
	Repo   *repository.DB
	Blob   *azureblob.Client
	SMS    SMSSender
}
