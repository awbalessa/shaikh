package rest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
	"github.com/awbalessa/shaikh/backend/internal/svc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v3"
	"github.com/google/uuid"
)

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	httplog.SetError(r.Context(), err)
	domainErr := dom.ToDomainError(err)

	var statusCode int
	switch domainErr.Code {
	case dom.CodeNotFound:
		statusCode = http.StatusNotFound
	case dom.CodeConflict:
		statusCode = http.StatusConflict
	case dom.CodeInvalidArgument:
		statusCode = http.StatusUnprocessableEntity
	case dom.CodeUnauthorized:
		statusCode = http.StatusUnauthorized
	case dom.CodeForbidden, dom.CodeOwnershipViolation:
		statusCode = http.StatusForbidden
	case dom.CodeTimeout:
		statusCode = http.StatusGatewayTimeout
	case dom.CodeUnavailable:
		statusCode = http.StatusServiceUnavailable
	default:
		statusCode = http.StatusInternalServerError
	}

	userMessage := "An unexpected error occurred. Please try again."
	switch domainErr.Code {
	case dom.CodeNotFound:
		userMessage = "The requested resource was not found."
	case dom.CodeInvalidArgument:
		userMessage = "Your request was invalid."
	case dom.CodeUnauthorized:
		userMessage = "You are not authorized to perform this action."
	case dom.CodeForbidden, dom.CodeOwnershipViolation:
		userMessage = "You do not have permission to access this resource."
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    domainErr.Code,
		"message": userMessage,
	})
}

func healthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}
}

func readyzHandler(hs *svc.HealthReadinessSvc) http.HandlerFunc {
	type payload struct {
		Status string            `json:"status"`
		Checks []svc.CheckResult `json:"checks"`
		TS     string            `json:"ts"`
		DurMS  int64             `json:"dur_ms"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		ready, results := hs.CheckReadiness(ctx)

		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		if !ready {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		_ = json.NewEncoder(w).Encode(payload{
			Status: map[bool]string{true: "ready", false: "unready"}[ready],
			Checks: results,
			TS:     time.Now().UTC().Format(time.RFC3339),
			DurMS:  time.Since(start).Milliseconds(),
		})
	}
}

func askHandler(ask *svc.AskSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Prompt string `json:"prompt"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, r, dom.NewTaggedError(dom.ErrInvalidInput, err))
			return
		}
		if strings.TrimSpace(body.Prompt) == "" {
			writeError(w, r, dom.NewTaggedError(dom.ErrInvalidInput, nil))
			return
		}

		userID, err := UserIDFromCtx(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}
		sessionID, err := SessionIDFromCtx(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}

		res, err := ask.Ask(r.Context(), body.Prompt, userID, sessionID)
		if err != nil {
			writeError(w, r, err)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		flusher, ok := w.(http.Flusher)
		if !ok {
			writeError(w, r, dom.NewTaggedError(dom.ErrInternal, nil))
			return
		}

		writeEvent := func(event string, payload any) bool {
			b, err := json.Marshal(payload)
			if err != nil {
				return false
			}
			if _, err = io.WriteString(w, "event: "+event+"\n"); err != nil {
				return false
			}
			if _, err = io.WriteString(w, "data: "); err != nil {
				return false
			}
			if _, err = w.Write(b); err != nil {
				return false
			}
			if _, err = io.WriteString(w, "\n\n"); err != nil {
				return false
			}
			flusher.Flush()
			return true
		}

		if !writeEvent("ready", struct{}{}) {
			return
		}

		for token, err := range res.Stream {
			select {
			case <-r.Context().Done():
				return
			default:
			}

			if err != nil {
				de := dom.ToDomainError(err)
				_ = writeEvent("error", map[string]string{"code": de.Code, "message": de.Message})
				return
			}

			if !writeEvent("token", map[string]string{"token": token}) {
				return
			}
		}

		_ = writeEvent("done", struct{}{})
	}
}

func registerHandler(u *svc.UserSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, r, dom.NewTaggedError(dom.ErrInvalidInput, err))
			return
		}
		body.Email = strings.ToLower(strings.TrimSpace(body.Email))

		user, err := u.Register(r.Context(), body.Email, body.Password)
		if err != nil {
			writeError(w, r, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"id":    user.ID,
			"email": user.Email,
		})
	}
}

func loginHandler(user *svc.UserSvc, au *svc.AuthSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, r, dom.NewTaggedError(dom.ErrInvalidInput, err))
			return
		}

		u, err := user.Login(r.Context(), body.Email, body.Password)
		if err != nil {
			writeError(w, r, err)
			return
		}

		acc, ref, err := au.IssueTokens(r.Context(), u)
		if err != nil {
			writeError(w, r, err)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "rt",
			Value:    ref,
			Path:     "/auth",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(60 * 24 * time.Hour),
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": acc,
			"token_type":   "Bearer",
			"expires_in":   int(au.JWT().TTL.Seconds()),
			"user": map[string]any{
				"id":    u.ID,
				"email": u.Email,
			},
		})
	}
}

func refreshHandler(au *svc.AuthSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rt, err := r.Cookie("rt")
		if err != nil || rt.Value == "" {
			writeError(w, r, dom.NewTaggedError(dom.ErrUnauthorized, err))
			return
		}

		acc, ref, err := au.Refresh(r.Context(), rt.Value)
		if err != nil {
			writeError(w, r, err)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "rt",
			Value:    ref,
			Path:     "/auth",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(60 * 24 * time.Hour),
		})

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": acc,
			"token_type":   "Bearer",
			"expires_in":   int(au.JWT().TTL.Seconds()),
		})
	}
}

func logoutHandler(au *svc.AuthSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("rt")
		if err == nil && cookie.Value != "" {
			_ = au.Revoke(r.Context(), cookie.Value)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "rt",
			Value:    "",
			Path:     "/auth",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Unix(0, 0),
		})

		w.WriteHeader(http.StatusNoContent)
	}
}

func logoutAllHandler(au *svc.AuthSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := UserIDFromCtx(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}
		if userID == uuid.Nil {
			writeError(w, r, dom.NewTaggedError(dom.ErrUnauthorized, nil))
			return
		}

		if err := au.RevokeAll(r.Context(), userID); err != nil {
			writeError(w, r, err)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "rt",
			Value:    "",
			Path:     "/auth",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Unix(0, 0),
		})

		w.WriteHeader(http.StatusNoContent)
	}
}

func createSessionHandler(sesh *svc.SessionSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := UserIDFromCtx(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}

		s, err := sesh.Create(r.Context(), userID)
		if err != nil {
			writeError(w, r, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"id":     s.ID,
			"userID": s.UserID,
		})
	}
}

func archiveSessionHandler(sesh *svc.SessionSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := UserIDFromCtx(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}
		sessionID, err := SessionIDFromCtx(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}

		var body struct {
			Archived bool `json:"archived"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, r, dom.NewTaggedError(dom.ErrInvalidInput, err))
			return
		}

		s, err := sesh.SetArchive(r.Context(), sessionID, userID, body.Archived)
		if err != nil {
			writeError(w, r, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"id":          s.ID,
			"archived_at": *s.ArchivedAt,
		})
	}
}

func deleteSessionHandler(sesh *svc.SessionSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := UserIDFromCtx(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}
		sessionID, err := SessionIDFromCtx(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}

		if err := sesh.Delete(r.Context(), sessionID, userID); err != nil {
			writeError(w, r, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func deleteUserHandler(u *svc.UserSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := UserIDFromCtx(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}

		if err := u.Delete(r.Context(), userID); err != nil {
			writeError(w, r, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func adminDeleteUserHandler(u *svc.UserSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "id")
		if userID == "" {
			writeError(w, r, dom.NewTaggedError(dom.ErrInvalidInput, nil))
			return
		}
		uid, err := uuid.Parse(userID)
		if err != nil {
			writeError(w, r, dom.NewTaggedError(dom.ErrInvalidInput, err))
			return
		}

		if err := u.Delete(r.Context(), uid); err != nil {
			writeError(w, r, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
