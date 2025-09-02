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
		var load struct {
			Prompt string `json:"prompt"`
		}

		if err := json.NewDecoder(r.Body).Decode(&load); err != nil || strings.TrimSpace(load.Prompt) == "" {
			http.Error(w, "invalid body: need {\"prompt\": \"...\"}", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		stream := ask.Ask(r.Context(), load.Prompt)

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
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}

		user, err := u.Register(r.Context(), body.Email, body.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

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
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}

		u, err := user.Login(r.Context(), body.Email, body.Password)
		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		token, err := tok.Sign(u.ID)
		if err != nil {
			http.Error(w, "cannot issue token", http.StatusInternalServerError)
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
