package svc

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
)

type HealthReadinessSvc struct {
	Providers []dom.Provider
}

type CheckResult struct {
	Name   string `json:"name"`
	Status string `json:"status"` // ok | down | skipped
	Error  string `json:"error,omitempty"`
}

func (s *HealthReadinessSvc) CheckReadiness(ctx context.Context) (bool, []CheckResult) {
	results := make([]CheckResult, len(s.Providers))

	var wg sync.WaitGroup
	wg.Add(len(s.Providers))

	for i, p := range s.Providers {
		i, p := i, p // capture
		go func() {
			defer wg.Done()

			cctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
			defer cancel()
			err := p.Ping(cctx)

			cr := CheckResult{Name: p.Name()}

			switch {
			case err == nil:
				cr.Status = "ok"
			case errors.Is(err, dom.ErrNotPingable):
				cr.Status = "skipped"
				cr.Error = dom.ErrNotPingable.Error()
			case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
				cr.Status = "timeout"
				cr.Error = err.Error()
			default:
				cr.Status = "down"
				cr.Error = err.Error()
			}

			results[i] = cr
		}()
	}

	wg.Wait()

	ready := true
	for _, cr := range results {
		if cr.Status == "down" {
			ready = false
			break
		}
	}

	return ready, results
}
