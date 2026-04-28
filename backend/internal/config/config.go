package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds application configuration loaded from the environment.
type Config struct {
	HTTPPort string

	DatabaseURL string

	JWTSecret string

	AzureStorageAccount   string
	AzureStorageKey       string
	AzureStorageContainer string

	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string

	// AllowSelfKycVerify enables POST /api/me/complete-kyc for demo/local environments only.
	AllowSelfKycVerify bool

	// JWTTTLHours controls access token lifetime (default 72).
	JWTTTLHours int
}

// Load reads configuration from the environment. If path is non-empty, loads .env from that path first.
func Load(dotEnvPath string) (*Config, error) {
	if dotEnvPath != "" {
		_ = godotenv.Load(dotEnvPath)
	} else {
		_ = godotenv.Load()
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	jwt := os.Getenv("JWT_SECRET")
	if jwt == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	selfKyc := os.Getenv("ALLOW_SELF_KYC_VERIFY") == "true" || os.Getenv("ALLOW_SELF_KYC_VERIFY") == "1"

	jwtHours := 72
	if v := os.Getenv("JWT_TTL_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 720 {
			jwtHours = n
		}
	}

	return &Config{
		HTTPPort:              port,
		DatabaseURL:           dbURL,
		JWTSecret:             jwt,
		AzureStorageAccount:   os.Getenv("AZURE_STORAGE_ACCOUNT"),
		AzureStorageKey:       os.Getenv("AZURE_STORAGE_KEY"),
		AzureStorageContainer: os.Getenv("AZURE_STORAGE_CONTAINER"),
		TwilioAccountSID:      os.Getenv("TWILIO_ACCOUNT_SID"),
		TwilioAuthToken:       os.Getenv("TWILIO_AUTH_TOKEN"),
		TwilioFromNumber:      os.Getenv("TWILIO_FROM_NUMBER"),
		AllowSelfKycVerify:    selfKyc,
		JWTTTLHours:           jwtHours,
	}, nil
}

// MustPort returns HTTP port as int for servers that need it.
func (c *Config) MustPort() int {
	n, err := strconv.Atoi(c.HTTPPort)
	if err != nil {
		return 8080
	}
	return n
}
