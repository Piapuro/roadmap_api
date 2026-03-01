package driver

import "os"

type SupabaseConfig struct {
	URL       string
	AnonKey   string
	JWTSecret string
}

func NewSupabaseConfig() *SupabaseConfig {
	return &SupabaseConfig{
		URL:       os.Getenv("SUPABASE_URL"),
		AnonKey:   os.Getenv("SUPABASE_ANON_KEY"),
		JWTSecret: os.Getenv("SUPABASE_JWT_SECRET"),
	}
}
