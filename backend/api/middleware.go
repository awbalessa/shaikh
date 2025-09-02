package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CtxKey string

const (
	CtxUserIDKey    CtxKey = "userID"
	CtxSessionIDKey CtxKey = "sessionID"
)

func UserAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := r.Header.Get("X-User-ID")
		if uid == "" {
			http.Error(w, "missing user id", http.StatusUnauthorized)
			return
		}
		userID, err := uuid.Parse(uid)
		if err != nil {
			http.Error(w, "invalid user id", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), CtxUserIDKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SessionAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sid := chi.URLParam(r, "sessionID")
		if sid == "" {
			http.Error(w, "missing session id", http.StatusUnauthorized)
			return
		}
		sessionID, err := uuid.Parse(sid)
		if err != nil {
			http.Error(w, "invalid session id", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), CtxSessionIDKey, sessionID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromCtx(ctx context.Context) (uuid.UUID, error) {
	v := ctx.Value(CtxUserIDKey)
	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("missing or invalid userID in context")
	}
	return id, nil
}

func SessionIDFromCtx(ctx context.Context) (uuid.UUID, error) {
	v := ctx.Value(CtxSessionIDKey)
	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("missing or invalid sessionID in context")
	}
	return id, nil
}
