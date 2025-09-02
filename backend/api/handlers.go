package api

import (
	"context"
	"encoding/json"
	"net/http"
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
