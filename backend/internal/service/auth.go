package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"carmanage/backend/internal/auth"
	"carmanage/backend/internal/httpx"
	"carmanage/backend/internal/models"

	"gorm.io/gorm"
)

// AuthService handles registration and login.
type AuthService struct {
	Deps
}

type RegisterInput struct {
	Email                string
	Password             string
	Role                 models.UserRole
	FullName             string
	AadhaarNumber        string
	PhoneNumber          string
	Address              string
	DrivingLicenseNumber *string
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*models.User, string, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	if in.Email == "" || len(in.Password) < 8 {
		return nil, "", httpx.WrapValidation("Email and a password of at least 8 characters are required.")
	}
	if in.Role != models.RoleCustomer && in.Role != models.RoleOwner {
		return nil, "", httpx.WrapValidation("Role must be CUSTOMER or OWNER.")
	}
	if strings.TrimSpace(in.FullName) == "" || strings.TrimSpace(in.AadhaarNumber) == "" ||
		strings.TrimSpace(in.PhoneNumber) == "" || strings.TrimSpace(in.Address) == "" {
		return nil, "", httpx.WrapValidation("Full name, Aadhaar number, phone number, and address are required.")
	}

	if _, err := s.Repo.GetUserByEmail(ctx, in.Email); err == nil {
		return nil, "", httpx.NewError(409, "EMAIL_IN_USE", "An account with this email already exists.")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, "", err
	}

	u := &models.User{
		Email:                in.Email,
		PasswordHash:         hash,
		Role:                 in.Role,
		FullName:             strings.TrimSpace(in.FullName),
		AadhaarNumber:        strings.TrimSpace(in.AadhaarNumber),
		PhoneNumber:          strings.TrimSpace(in.PhoneNumber),
		Address:              strings.TrimSpace(in.Address),
		DrivingLicenseNumber: trimPtr(in.DrivingLicenseNumber),
		IsKYCVerified:        false,
	}
	if err := s.Repo.CreateUser(ctx, u); err != nil {
		return nil, "", err
	}

	token, err := s.issueToken(u)
	if err != nil {
		return nil, "", err
	}
	return u, token, nil
}

type LoginInput struct {
	Email    string
	Password string
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*models.User, string, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	u, err := s.Repo.GetUserByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", httpx.NewError(401, "INVALID_CREDENTIALS", "The email or password you entered is incorrect.")
		}
		return nil, "", err
	}
	if !auth.CheckPassword(u.PasswordHash, in.Password) {
		return nil, "", httpx.NewError(401, "INVALID_CREDENTIALS", "The email or password you entered is incorrect.")
	}
	token, err := s.issueToken(u)
	if err != nil {
		return nil, "", err
	}
	return u, token, nil
}

func (s *AuthService) issueToken(u *models.User) (string, error) {
	ttl := time.Duration(s.Config.JWTTTLHours) * time.Hour
	if ttl <= 0 {
		ttl = 72 * time.Hour
	}
	return auth.IssueAccessToken(s.Config.JWTSecret, u.ID, string(u.Role), u.Email, ttl)
}

func trimPtr(p *string) *string {
	if p == nil {
		return nil
	}
	t := strings.TrimSpace(*p)
	if t == "" {
		return nil
	}
	return &t
}
