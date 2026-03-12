package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AppEnv            string
	Port              string
	DatabaseURL       string
	SupabaseURL       string
	SupabaseAnonKey   string
	SupabaseJWTSecret string
	GeminiAPIKey      string
	WebhookSecret     string
	CORSAllowOrigins  string
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// Load は起動時に全環境変数を読み込み、必須項目が未設定なら即エラーを返す。
func Load() (*Config, error) {
	var missing []string

	require := func(key string) string {
		v := os.Getenv(key)
		if v == "" {
			missing = append(missing, key)
		}
		return v
	}

	withDefault := func(key, defaultVal string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return defaultVal
	}

	cfg := &Config{
		AppEnv:            withDefault("APP_ENV", "development"),
		Port:              withDefault("PORT", "8080"),
		CORSAllowOrigins:  withDefault("CORS_ALLOW_ORIGINS", "http://localhost:3000"),
		DatabaseURL:       require("DATABASE_URL"),
		SupabaseURL:       require("SUPABASE_URL"),
		SupabaseAnonKey:   require("SUPABASE_ANON_KEY"),
		SupabaseJWTSecret: require("SUPABASE_JWT_SECRET"),
		GeminiAPIKey:      require("GEMINI_API_KEY"),
		WebhookSecret:     require("WEBHOOK_SECRET"),
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}
