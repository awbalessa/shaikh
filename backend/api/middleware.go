package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/awbalessa/shaikh/backend/internal/svc"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTValidator struct {
	Secret   []byte
	Issuer   string
	Audience string
}

func NewJWTValidator() *JWTValidator {
	return &JWTValidator{
		Secret:   []byte(os.Getenv("JWT_SECRET")),
		Issuer:   "shaikh-api",
		Audience: "shaikh-api",
	}
}

func (v *JWTValidator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}

		raw := strings.TrimPrefix(h, "Bearer ")

		token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected alg")
			}
			return v.Secret, nil
		}, jwt.WithAudience(v.Audience), jwt.WithIssuer(v.Issuer))
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "expired or invalid token", http.StatusUnauthorized)
			return
		}

		claims, _ := token.Claims.(jwt.MapClaims)
		sub, _ := claims["sub"].(string)
		uid, err := uuid.Parse(sub)
		if err != nil {
			http.Error(w, "bad jwt sub", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), svc.CtxUserIDKey, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SessionAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sid := chi.URLParam(r, "sessionID")
		if sid == "" {
			http.Error(w, "missing session id", http.StatusBadRequest)
			return
		}
		sessionID, err := uuid.Parse(sid)
		if err != nil {
			http.Error(w, "invalid session id", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), svc.CtxSessionIDKey, sessionID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
