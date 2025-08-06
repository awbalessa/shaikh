package worker

import (
	"context"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/agent"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Worker interface {
	start()
	process(ctx context.Context, msg jetstream.Msg) error
}

const (
	syncerDurableName         string        = "agent-context-syncer"
	syncerAckTime             time.Duration = 3 * time.Minute
	syncerMaxDeliveryAttempts int           = 5
)

type syncer struct {
	name      string
	cons      jetstream.Consumer
	lastFlush time.Time
	buffer    []agent.Interaction
}

func buildSyncer(ctx context.Context, stream jetstream.JetStream) (*syncer, error) {
	cons, err := stream.CreateOrUpdateConsumer(ctx, agent.AgentStream, jetstream.ConsumerConfig{
		Durable:       syncerDurableName,
		FilterSubject: agent.SyncerSubject,
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       syncerAckTime,
		MaxDeliver:    syncerMaxDeliveryAttempts,
		ReplayPolicy:  jetstream.ReplayInstantPolicy,
	})
	if err != nil {
		return nil, err
	}

	return &syncer{
		name:      syncerDurableName,
		cons:      cons,
		lastFlush: time.Time{},
		buffer:    nil,
	}, nil
}

func (s *syncer) process(ctx context.Context, msg *nats.Msg) error {
	s.buffer = append(s.buffer, msg)

	if time.Since(s.lastFlush) >= agent.SyncIdleTime {
		if err := s.flush(ctx); err != nil {
			return err
		}
	}

	if len(s.buffer) >= agent.SyncMaxBatchSize {
		if err := s.flush(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *syncer) flush(ctx context.Context) error {
	if len(s.buffer) == 0 {
		return nil
	}

	return nil
}
