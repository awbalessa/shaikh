package ai

import (
	"context"
	"io"
	"time"
)

type FakeModel struct{}

func (FakeModel) ID() string       { return "fake-1" }
func (FakeModel) Provider() string { return "fake" }


func (FakeModel) Generate(ctx context.Context, call CallOptions) (GenerateResult, error) {
	return GenerateResult{
		Contents: []Content{
			TextContent{Text: "This is a fake generated response."},
		},
		Usage: &Usage{
			InputTokens:  10,
			OutputTokens: 10,
			TotalTokens:  20,
		},
	}, nil
}

func (FakeModel) Stream(ctx context.Context, call CallOptions) (StreamResult, error) {
	return newFakeStream(ctx, call, time.Millisecond*200), nil
}

type fakeStream struct {
	ctx    context.Context
	call   CallOptions
	step   int
	delay  time.Duration
}

func newFakeStream(ctx context.Context, call CallOptions, delay time.Duration) *fakeStream {
	return &fakeStream{ctx: ctx, call: call, delay: delay}
}

func (s *fakeStream) Recv() (Event, error) {
	select {
	case <-s.ctx.Done():
		return Event{}, s.ctx.Err()
	default:
	}

	time.Sleep(s.delay)

	switch s.step {
	case 0:
		s.step++
		return Event{Type: EventStreamStart}, nil

	case 1:
		s.step++
		return Event{
			Type: EventTextStart,
			ID:   "msg_1",
		}, nil

	case 2:
		s.step++
		return Event{
			Type:  EventTextDelta,
			ID:    "msg_1",
			Delta: "Hello",
		}, nil

	case 3:
		s.step++
		return Event{
			Type:  EventTextDelta,
			ID:    "msg_1",
			Delta: " world",
		}, nil

	case 4:
		s.step++
		return Event{
			Type: EventTextEnd,
			ID:   "msg_1",
		}, nil

	case 5:
		s.step++
		return Event{
			Type:   EventFinish,
			Reason: FinishReasonStop,
			Usage: &Usage{
				InputTokens:  10,
				OutputTokens: 5,
				TotalTokens:  15,
			},
		}, nil

	default:
		return Event{}, io.EOF
	}
}

func (s *fakeStream) Close() error {
	return nil
}
