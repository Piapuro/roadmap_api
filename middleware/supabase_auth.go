package middleware

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const ContextKeyUserID = "user_id"

type SupabaseAuth struct {
	jwtSecret []byte
	issuer    string
	ecKeys    map[string]*ecdsa.PublicKey // kid → 公開鍵（ES256用）
}

// jwk は JWKS エンドポイントから取得する1つの鍵を表す
type jwk struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

func NewSupabaseAuth(jwtSecret string, issuer string) *SupabaseAuth {
	auth := &SupabaseAuth{
		jwtSecret: []byte(jwtSecret),
		issuer:    issuer,
		ecKeys:    make(map[string]*ecdsa.PublicKey),
	}
	// 起動時にJWKSを取得してキャッシュする（ES256対応）
	if issuer != "" {
		if err := auth.fetchJWKS(); err != nil {
			log.Printf("[SupabaseAuth] JWKS fetch failed (ES256 tokens will be rejected): %v", err)
		}
	}
	return auth
}

// fetchJWKS は Supabase の JWKS エンドポイントから公開鍵を取得してキャッシュする
func (m *SupabaseAuth) fetchJWKS() error {
	url := strings.TrimSuffix(m.issuer, "/") + "/.well-known/jwks.json"
	resp, err := http.Get(url) //nolint:noctx // 起動時の1回限りの呼び出し
	if err != nil {
		return fmt.Errorf("get jwks: %w", err)
	}
	defer resp.Body.Close()

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("decode jwks: %w", err)
	}

	for _, key := range jwks.Keys {
		if key.Kty != "EC" {
			continue
		}
		pub, err := parseECPublicKey(key)
		if err != nil {
			log.Printf("[SupabaseAuth] skip key %s: %v", key.Kid, err)
			continue
		}
		m.ecKeys[key.Kid] = pub
	}
	return nil
}

func parseECPublicKey(key jwk) (*ecdsa.PublicKey, error) {
	xBytes, err := base64.RawURLEncoding.DecodeString(key.X)
	if err != nil {
		return nil, fmt.Errorf("decode x: %w", err)
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(key.Y)
	if err != nil {
		return nil, fmt.Errorf("decode y: %w", err)
	}

	var curve elliptic.Curve
	switch key.Crv {
	case "P-256":
		curve = elliptic.P256()
	case "P-384":
		curve = elliptic.P384()
	case "P-521":
		curve = elliptic.P521()
	default:
		return nil, fmt.Errorf("unsupported curve: %s", key.Crv)
	}

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}

type supabaseClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// Verify validates the Supabase JWT and sets claims["sub"] as user_id in context.
// HS256（旧プロジェクト）と ES256（新プロジェクト）の両方に対応する。
func (m *SupabaseAuth) Verify(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &supabaseClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			switch t.Method.(type) {
			case *jwt.SigningMethodHMAC:
				return m.jwtSecret, nil
			case *jwt.SigningMethodECDSA:
				kid, ok := t.Header["kid"].(string)
				if !ok {
					log.Printf("[SupabaseAuth] ES256 token has no kid header")
					return nil, echo.ErrUnauthorized
				}
				pub, found := m.ecKeys[kid]
				if !found {
					log.Printf("[SupabaseAuth] kid=%s not found in cache (cached keys: %d)", kid, len(m.ecKeys))
					return nil, echo.ErrUnauthorized
				}
				return pub, nil
			default:
				log.Printf("[SupabaseAuth] unsupported alg: %s", t.Method.Alg())
				return nil, echo.ErrUnauthorized
			}
		}, jwt.WithExpirationRequired(), jwt.WithAudience("authenticated"))
		if err != nil || !token.Valid {
			log.Printf("[SupabaseAuth] token parse failed: %v", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		if m.issuer != "" && claims.Issuer != m.issuer {
			log.Printf("[SupabaseAuth] issuer mismatch: got=%q want=%q", claims.Issuer, m.issuer)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		sub, err := claims.GetSubject()
		if err != nil || sub == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		c.Set(ContextKeyUserID, sub)
		return next(c)
	}
}
