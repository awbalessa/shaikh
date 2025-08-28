package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	syncerDurableName         string        = "fix-syncer"
	syncerAckTime             time.Duration = 3 * time.Minute
	syncerMaxDeliveryAttempts int           = 5
)

type syncer struct {
	name      string
	cons      jetstream.Consumer
	log       *slog.Logger
	lastFlush time.Time
	buffer    []jetstream.Msg
}

func BuildSyncer(
	ctx context.Context,
	stream jetstream.JetStream,
) (*syncer, error) {
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

	logger := slog.Default().With(
		"component", "worker-group",
		slog.String("worker", syncerDurableName),
		slog.String("subject", agent.SyncerSubject),
		slog.String("ack_time", syncerAckTime.String()),
	)

	return &syncer{
		name:      syncerDurableName,
		cons:      cons,
		log:       logger,
		lastFlush: time.Time{},
		buffer:    nil,
	}, nil
}

const (
	SyncerDurableName string        = "syncer"
	SyncerSubject     string        = dom.ContextStreamSubject + ".sync"
	SyncIdleTime      time.Duration = 2 * time.Minute
	SyncMaxBatchSize  int           = 5
)

type Syncer struct {
	Cons       dom.PubSubConsumer
	LastFlush  time.Time
	Buffer     []dom.PubMsg
	UnitOfWork dom.UnitOfWork
}

func (s *Syncer) Start(ctx context.Context) error {
	msgs, err := s.Cons.Messages(ctx)
	if err != nil {
		return fmt.Errorf("syncer failed: %w", err)
	}

	for m := range msgs {
		if err = s.Process(ctx, m); err != nil {
			return fmt.Errorf("syncer failed: %w", err)
		}
	}
	return nil
}

func (s *Syncer) Process(ctx context.Context, msg dom.PubMsg) error {
	s.Buffer = append(s.Buffer, msg)

	if len(s.Buffer) >= SyncMaxBatchSize {
		if err := s.flush(ctx); err != nil {
			return fmt.Errorf("failed to process message: %w", err)
		}
	}

	if time.Since(s.LastFlush) >= SyncIdleTime {
		if err := s.flush(ctx); err != nil {
			return fmt.Errorf("failed to process message: %w", err)
		}
	}

	return nil
}

func (s *Syncer) flush(ctx context.Context) error {
	tx, err := s.UnitOfWork.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to flush: %w", err)
	}
	defer tx.Rollback(ctx)

	loads := make([]SyncPayloadDTO, len(s.Buffer))
	for i, m := range s.Buffer {
		if err := json.Unmarshal(m.Data(), &loads[i]); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		if err := s.persistMessages(ctx, tx, loads[i]); err != nil {
			return fmt.Errorf("failed to create messages from interaction: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to flush: %w", err)
	}

	for _, m := range s.Buffer {
		if err := m.Ack(); err != nil {
			return fmt.Errorf("failed to flush: %w", err)
		}
	}

	s.Buffer = nil
	s.LastFlush = time.Now()
	return nil
}

func (s *Syncer) persistMessages(
	ctx context.Context,
	tx dom.Tx,
	load SyncPayloadDTO,
) error {
	var mr dom.MessageRepo
	if err := tx.Get(&mr); err != nil {
		return fmt.Errorf("failed to persist messages: %w", err)
	}

	_, err := mr.CreateMessage(ctx context.Context, msg dom.Message)
}
