package driver

type SupabaseConfig struct {
	URL       string
	AnonKey   string
	JWTSecret string
}

func NewSupabaseConfig(url, anonKey, jwtSecret string) *SupabaseConfig {
	return &SupabaseConfig{
		URL:       url,
		AnonKey:   anonKey,
		JWTSecret: jwtSecret,
	}
}
