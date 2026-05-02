package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Environment names must match files under backend/config/<name>.yaml
const (
	EnvDev  = "dev"
	EnvStag = "stag"
	EnvProd = "prod"
)

// Config holds application configuration from backend/config/<APP_ENV>.yaml
// after reading APP_ENV from backend/.env (or the process environment).
type Config struct {
	Environment string

	HTTPPort string
	GinMode  string

	DatabaseURL string

	JWTSecret string

	AzureStorageAccount   string
	AzureStorageKey       string
	AzureStorageContainer string

	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string

	AllowSelfKycVerify bool
	JWTTTLHours        int

	// Platform commission as percent of agreed rental (charged to customer on top of base; deducted from owner payout).
	CustomerCommissionPercent float64
	OwnerCommissionPercent    float64
	// GST percent on customer subtotal (agreed rental + customer platform fee) and on owner agreed rental; see booking payment math.
	GstPercentOnCommission float64
}

type yamlFile struct {
	HTTPPort string `yaml:"http_port"`
	GinMode  string `yaml:"gin_mode"`

	DatabaseURL string `yaml:"database_url"`

	JWTSecret   string `yaml:"jwt_secret"`
	JWTTTLHours int    `yaml:"jwt_ttl_hours"`

	AllowSelfKycVerify *bool `yaml:"allow_self_kyc_verify"`

	Azure struct {
		StorageAccount   string `yaml:"storage_account"`
		StorageKey       string `yaml:"storage_key"`
		StorageContainer string `yaml:"storage_container"`
	} `yaml:"azure"`

	Twilio struct {
		AccountSID string `yaml:"account_sid"`
		AuthToken  string `yaml:"auth_token"`
		FromNumber string `yaml:"from_number"`
	} `yaml:"twilio"`

	Payments struct {
		CustomerCommissionPercent float64 `yaml:"customer_commission_percent"`
		OwnerCommissionPercent    float64 `yaml:"owner_commission_percent"`
		GstPercentOnCommission    float64 `yaml:"gst_percent_on_commission"`
	} `yaml:"payments"`
}

// Load reads backend/.env for APP_ENV only, then loads backend/config/<APP_ENV>.yaml.
// backendRoot is the directory that contains .env and the config/ folder (the module root).
func Load(backendRoot string) (*Config, error) {
	envPath := filepath.Join(backendRoot, ".env")
	if _, err := os.Stat(envPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("missing %s: create it with APP_ENV=dev|stag|prod", envPath)
		}
		return nil, fmt.Errorf("stat .env: %w", err)
	}
	if err := godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	rawEnv := strings.TrimSpace(strings.ToLower(os.Getenv("APP_ENV")))
	if rawEnv == "" {
		return nil, fmt.Errorf("APP_ENV is required in .env (use dev, stag, or prod)")
	}
	switch rawEnv {
	case EnvDev, EnvStag, EnvProd:
	default:
		return nil, fmt.Errorf("APP_ENV must be one of dev, stag, prod; got %q", rawEnv)
	}

	yamlPath := filepath.Join(backendRoot, "config", rawEnv+".yaml")
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", yamlPath, err)
	}

	var y yamlFile
	if err := yaml.Unmarshal(data, &y); err != nil {
		return nil, fmt.Errorf("parse %s: %w", yamlPath, err)
	}

	port := strings.TrimSpace(y.HTTPPort)
	if port == "" {
		port = "8080"
	}
	ginMode := strings.TrimSpace(y.GinMode)
	if ginMode == "" {
		ginMode = "debug"
	}

	dbURL := strings.TrimSpace(y.DatabaseURL)
	if dbURL == "" {
		return nil, fmt.Errorf("database_url is required in %s", yamlPath)
	}

	jwt := strings.TrimSpace(y.JWTSecret)
	if jwt == "" {
		return nil, fmt.Errorf("jwt_secret is required in %s", yamlPath)
	}

	selfKyc := false
	if y.AllowSelfKycVerify != nil {
		selfKyc = *y.AllowSelfKycVerify
	}

	jwtHours := y.JWTTTLHours
	if jwtHours <= 0 {
		jwtHours = 72
	}
	if jwtHours > 720 {
		jwtHours = 720
	}

	custComm := y.Payments.CustomerCommissionPercent
	ownerComm := y.Payments.OwnerCommissionPercent
	if custComm < 0 {
		custComm = 0
	}
	if ownerComm < 0 {
		ownerComm = 0
	}
	if custComm == 0 && ownerComm == 0 {
		custComm, ownerComm = 2, 1.5
	}

	gstOnComm := y.Payments.GstPercentOnCommission
	if gstOnComm < 0 {
		gstOnComm = 0
	}
	if gstOnComm == 0 {
		gstOnComm = 18
	}

	// Optional env overrides (backend/.env) so secrets are not committed in YAML.
	if v := strings.TrimSpace(os.Getenv("DATABASE_URL")); v != "" {
		dbURL = v
	}
	if v := strings.TrimSpace(os.Getenv("JWT_SECRET")); v != "" {
		jwt = v
	}
	azureAccount := strings.TrimSpace(y.Azure.StorageAccount)
	azureKey := strings.TrimSpace(y.Azure.StorageKey)
	azureContainer := strings.TrimSpace(y.Azure.StorageContainer)
	if v := strings.TrimSpace(os.Getenv("AZURE_STORAGE_ACCOUNT")); v != "" {
		azureAccount = v
	}
	if v := strings.TrimSpace(os.Getenv("AZURE_STORAGE_KEY")); v != "" {
		azureKey = v
	}
	if v := strings.TrimSpace(os.Getenv("AZURE_STORAGE_CONTAINER")); v != "" {
		azureContainer = v
	}

	return &Config{
		Environment:               rawEnv,
		HTTPPort:                  port,
		GinMode:                   ginMode,
		DatabaseURL:               dbURL,
		JWTSecret:                 jwt,
		AzureStorageAccount:       azureAccount,
		AzureStorageKey:           azureKey,
		AzureStorageContainer:     azureContainer,
		TwilioAccountSID:          strings.TrimSpace(y.Twilio.AccountSID),
		TwilioAuthToken:           strings.TrimSpace(y.Twilio.AuthToken),
		TwilioFromNumber:          strings.TrimSpace(y.Twilio.FromNumber),
		AllowSelfKycVerify:        selfKyc,
		JWTTTLHours:               jwtHours,
		CustomerCommissionPercent: custComm,
		OwnerCommissionPercent:    ownerComm,
		GstPercentOnCommission:    gstOnComm,
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
