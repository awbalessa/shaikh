package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/svc"
)

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
		Status string            `json:"status"` // ready | unready
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

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Prompt) == "" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		userID, err := svc.UserIDFromCtx(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sessionID, err := svc.SessionIDFromCtx(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		stream := ask.Ask(r.Context(), body.Prompt, userID, sessionID)

		io.WriteString(w, "event: ready\n")
		io.WriteString(w, "data: {}\n\n")
		flusher.Flush()

		for token, err := range stream {
			select {
			case <-r.Context().Done():
				return
			default:
			}

			if err != nil {
				payload, _ := json.Marshal(map[string]string{"error": err.Error()})
				io.WriteString(w, "event: error\n")
				io.WriteString(w, "data: ")
				w.Write(payload)
				io.WriteString(w, "\n\n")
				flusher.Flush()
				return
			}

			payload, _ := json.Marshal(map[string]string{"token": token})
			io.WriteString(w, "event: token\n")
			io.WriteString(w, "data: ")
			w.Write(payload)
			io.WriteString(w, "\n\n")
			flusher.Flush()
		}

		io.WriteString(w, "event: done\n")
		io.WriteString(w, "data: {}\n\n")
		flusher.Flush()
	}
}

func registerHandler(u *svc.UserSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		body.Email = strings.ToLower(strings.TrimSpace(body.Email))

		user, err := u.Register(r.Context(), body.Email, body.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
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

func loginHandler(user *svc.UserSvc, tok *svc.JWTIssuer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		u, err := user.Login(r.Context(), body.Email, body.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		token, err := tok.Sign(u.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": token,
			"token_type":   "Bearer",
			"expires_in":   int(tok.TTL.Seconds()),
			"user": map[string]any{
				"id":    u.ID,
				"email": u.Email,
			},
		})
	}
}

func createSessionHandler(sesh *svc.SessionSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := svc.UserIDFromCtx(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s, err := sesh.Create(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
		userID, err := svc.UserIDFromCtx(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sessionID, err := svc.SessionIDFromCtx(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var body struct {
			Archived bool `json:"archived"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s, err := sesh.SetArchive(r.Context(), sessionID, userID, body.Archived)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
		userID, err := svc.UserIDFromCtx(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sessionID, err := svc.SessionIDFromCtx(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := sesh.Delete(r.Context(), sessionID, userID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func deleteUserHandler(u *svc.UserSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := svc.UserIDFromCtx(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := u.Delete(r.Context(), userID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
