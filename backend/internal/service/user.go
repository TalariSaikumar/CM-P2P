package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"carmanage/backend/internal/httpx"
	"carmanage/backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserService handles profile and KYC helper flows.
type UserService struct {
	Deps
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u, err := s.Repo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

type UpdateProfileInput struct {
	FullName             *string
	AadhaarNumber        *string
	PhoneNumber          *string
	Address              *string
	DrivingLicenseNumber *string
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, in UpdateProfileInput) (*models.User, error) {
	u, err := s.Repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	if in.FullName != nil {
		v := strings.TrimSpace(*in.FullName)
		if v != "" {
			u.FullName = v
		}
	}
	if in.AadhaarNumber != nil {
		v := strings.TrimSpace(*in.AadhaarNumber)
		if v != "" {
			u.AadhaarNumber = v
		}
	}
	if in.PhoneNumber != nil {
		v := strings.TrimSpace(*in.PhoneNumber)
		if v != "" {
			u.PhoneNumber = v
		}
	}
	if in.Address != nil {
		v := strings.TrimSpace(*in.Address)
		if v != "" {
			u.Address = v
		}
	}
	if in.DrivingLicenseNumber != nil {
		v := strings.TrimSpace(*in.DrivingLicenseNumber)
		if v == "" {
			u.DrivingLicenseNumber = nil
		} else {
			u.DrivingLicenseNumber = &v
		}
	}
	if err := s.Repo.UpdateUser(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// CompleteKYC marks the user verified when explicitly allowed by server configuration (demo/local only).
func (s *UserService) CompleteKYC(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	if !s.Config.AllowSelfKycVerify {
		return nil, httpx.NewError(403, "KYC_FLOW_DISABLED", "Self-service verification is disabled on this server.")
	}
	u, err := s.Repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	if strings.TrimSpace(u.FullName) == "" || strings.TrimSpace(u.AadhaarNumber) == "" ||
		strings.TrimSpace(u.PhoneNumber) == "" || strings.TrimSpace(u.Address) == "" {
		return nil, httpx.WrapValidation("Complete your profile details before requesting verification.")
	}
	u.IsKYCVerified = true
	if err := s.Repo.UpdateUser(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// UploadKYCAttachment streams a document into blob storage and records metadata.
func (s *UserService) UploadKYCAttachment(ctx context.Context, userID uuid.UUID, kind models.KYCAttachmentKind, filename string, contentType string, body io.Reader) (*models.KYCAttachment, error) {
	if s.Blob == nil {
		return nil, httpx.ErrStorage
	}
	if body == nil {
		return nil, httpx.WrapValidation("File payload is required.")
	}

	safeName := strings.ReplaceAll(strings.ToLower(filename), "..", "")
	if safeName == "" {
		safeName = "document"
	}
	blobPath := fmt.Sprintf("kyc/%s/%s-%d", userID.String(), safeName, time.Now().UnixNano())
	url, err := s.Blob.Upload(ctx, blobPath, body, contentType)
	if err != nil {
		return nil, fmt.Errorf("upload kyc: %w", err)
	}

	a := &models.KYCAttachment{
		UserID:   userID,
		Kind:     kind,
		BlobPath: blobPath,
		BlobURL:  url,
	}
	if err := s.Repo.CreateKYCAttachment(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}
