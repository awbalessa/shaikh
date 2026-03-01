package fake

import (
	"context"
	"io"
	"time"

	"github.com/awbalessa/shaikh/api/internal/app/ai"
)

type Model struct{}

func (Model) ID() string       { return "fake-1" }
func (Model) Provider() string { return "fake" }

func (Model) Generate(ctx context.Context, call ai.CallOptions) (ai.GenerateResult, error) {
	return ai.GenerateResult{
		Contents: []ai.Content{
			ai.TextContent{Text: "This is a fake generated response."},
		},
		Usage: ai.Usage{
			InputTokens:  10,
			OutputTokens: 10,
			TotalTokens:  20,
		},
	}, nil
}

func (Model) Stream(ctx context.Context, call ai.CallOptions) (ai.StreamResult, error) {
	return ai.StreamResult{Stream: newStream(ctx, call, time.Millisecond*200)}, nil
}

type stream struct {
	ctx   context.Context
	call  ai.CallOptions
	step  int
	delay time.Duration
}

func newStream(ctx context.Context, call ai.CallOptions, delay time.Duration) *stream {
	return &stream{ctx: ctx, call: call, delay: delay}
}

func (s *stream) Recv() (ai.Event, error) {
	select {
	case <-s.ctx.Done():
		return ai.Event{}, s.ctx.Err()
	default:
	}

	time.Sleep(s.delay)

	switch s.step {
	case 0:
		s.step++
		return ai.Event{Type: ai.EventStreamStart}, nil

	case 1:
		s.step++
		return ai.Event{
			Type: ai.EventTextStart,
			ID:   "msg_1",
		}, nil

	case 2:
		s.step++
		return ai.Event{
			Type:  ai.EventTextDelta,
			ID:    "msg_1",
			Delta: "Hello",
		}, nil

	case 3:
		s.step++
		return ai.Event{
			Type:  ai.EventTextDelta,
			ID:    "msg_1",
			Delta: " world",
		}, nil

	case 4:
		s.step++
		return ai.Event{
			Type: ai.EventTextEnd,
			ID:   "msg_1",
		}, nil

	case 5:
		s.step++
		return ai.Event{
			Type:   ai.EventFinish,
			Reason: ai.FinishReasonStop,
			Usage: &ai.Usage{
				InputTokens:  10,
				OutputTokens: 5,
				TotalTokens:  15,
			},
		}, nil

	default:
		return ai.Event{}, io.EOF
	}
}

func (s *stream) Close() error {
	return nil
}
