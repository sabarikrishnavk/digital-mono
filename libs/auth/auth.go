package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWTAuthenticator handles JWT creation and validation.
type JWTAuthenticator struct {
	jwtSecret []byte
}

// NewJWTAuthenticator creates a new JWTAuthenticator.
func NewJWTAuthenticator(secret string) *JWTAuthenticator {
	return &JWTAuthenticator{jwtSecret: []byte(secret)}
}

// Claims defines the JWT claims.
type Claims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

type contextKey string

const ClaimsContextKey contextKey = "jwtClaims"

// GenerateToken creates a new JWT token.
func (a *JWTAuthenticator) GenerateToken(userID string, roles []string, duration time.Duration) (string, error) {
	expirationTime := time.Now().Add(duration)
	claims := &Claims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "user-service", // Or your issuer name
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(a.jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Middleware is an HTTP middleware for validating JWT tokens.
func (a *JWTAuthenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Authorization header format must be Bearer {token}", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return a.jwtSecret, nil
		})

		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				http.Error(w, "Token has expired", http.StatusUnauthorized)
			} else {
				http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			}
			return
		}

		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetClaimsFromContext retrieves the Claims struct from the context.
// It returns the Claims and a boolean indicating if the claims were found.
func GetClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	if !ok || claims == nil {
		return nil, false
	}
	return claims, true
}