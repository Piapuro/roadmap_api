package driver

import (
	"fmt"
	"os"
)

type SupabaseConfig struct {
	URL       string
	AnonKey   string
	JWTSecret string
}

func NewSupabaseConfig() (*SupabaseConfig, error) {
	secret := os.Getenv("SUPABASE_JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("SUPABASE_JWT_SECRET is not set")
	}
	return &SupabaseConfig{
		URL:       os.Getenv("SUPABASE_URL"),
		AnonKey:   os.Getenv("SUPABASE_ANON_KEY"),
		JWTSecret: secret,
	}, nil
}
