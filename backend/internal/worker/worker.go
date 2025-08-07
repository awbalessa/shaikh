package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/agent"
	"github.com/awbalessa/shaikh/backend/internal/database"
	"github.com/awbalessa/shaikh/backend/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nats-io/nats.go/jetstream"
)

type WorkerGroup struct {
	workers []Worker
}

func (wg *WorkerGroup) Add(w Worker) {
	wg.workers = append(wg.workers, w)
}

func (wg *WorkerGroup) StartAll(ctx context.Context, cancel context.CancelFunc) {
	for _, w := range wg.workers {
		go func(w Worker) {
			if err := w.start(ctx); err != nil {
				w.logger().With(
					"err", err,
				).ErrorContext(ctx, "worker exited")
				cancel()
			}
		}(w)
	}
}

type Worker interface {
	start(ctx context.Context) error
	process(ctx context.Context, msg jetstream.Msg) error
	logger() *slog.Logger
}

const (
	syncerDurableName         string        = "agent-context-syncer"
	syncerAckTime             time.Duration = 3 * time.Minute
	syncerMaxDeliveryAttempts int           = 5
)

type syncer struct {
	name      string
	cons      jetstream.Consumer
	store     *store.Store
	log       *slog.Logger
	lastFlush time.Time
	buffer    []jetstream.Msg
}

func buildSyncer(
	ctx context.Context,
	stream jetstream.JetStream,
	store *store.Store,
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
		store:     store,
		log:       logger,
		lastFlush: time.Time{},
		buffer:    nil,
	}, nil
}

func (s *syncer) logger() *slog.Logger {
	return s.log
}

func (s *syncer) start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			batch, err := s.cons.Fetch(
				agent.SyncMaxBatchSize,
			)
			if err != nil {
				return fmt.Errorf("syncer failed: %w", err)
			}

			for m := range batch.Messages() {
				if err := s.process(ctx, m); err != nil {
					return fmt.Errorf("syncer failed: %w", err)
				}
			}

			if err := batch.Error(); err != nil {
				return fmt.Errorf("syncer failed: %w", err)
			}
		}
	}
}

func (s *syncer) process(ctx context.Context, msg jetstream.Msg) error {
	s.buffer = append(s.buffer, msg)

	if len(s.buffer) >= agent.SyncMaxBatchSize {
		if err := s.flush(ctx); err != nil {
			return err
		}
	}

	if time.Since(s.lastFlush) >= agent.SyncIdleTime {
		if err := s.flush(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *syncer) flush(ctx context.Context) error {
	tx, err := s.store.Pg.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}
	defer tx.Rollback(ctx)

	payloads := make([]agent.SyncPayload, len(s.buffer))
	for i, m := range s.buffer {
		if err := json.Unmarshal(m.Data(), &payloads[i]); err != nil {
			return fmt.Errorf("failed to flush buffer: %w", err)
		}

		if err := s.createMessagesFromInteraction(ctx, tx, payloads[i]); err != nil {
			return fmt.Errorf("failed to flush buffer: %w", err)
		}

		if err := m.Ack(); err != nil {
			return fmt.Errorf("failed to flush buffer: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}

	s.buffer = nil
	s.lastFlush = time.Now()
	return nil
}

func (s *syncer) createMessagesFromInteraction(
	ctx context.Context,
	tx pgx.Tx,
	load agent.SyncPayload,
) error {
	pgSessionId := pgtype.UUID{
		Valid: true,
		Bytes: load.SessionID,
	}

	pgUserId := pgtype.UUID{
		Valid: true,
		Bytes: load.UserID,
	}
	_, err := s.store.Pg.CreateMessageTx(ctx, tx, database.CreateMessageParams{
		SessionID: pgSessionId,
		UserID:    pgUserId,
		Role:      database.MessagesRoleUser,
		Content:   load.Interaction.Input.UserInput.Text,
		Turn:      int32(load.Interaction.TurnNumber),
	})
	if err != nil {
		return fmt.Errorf("failed to create messages from interaction: %w", err)
	}
	_, err = s.store.Pg.CreateMessageTx(ctx, tx, database.CreateMessageParams{
		SessionID: pgSessionId,
		UserID:    pgUserId,
		Role:      database.MessagesRoleFunction,
		FunctionName: pgtype.Text{
			Valid:  true,
			String: string(load.Interaction.Input.FunctionName),
		},
		Content: load.Interaction.Input.FunctionResponse.Text,
		Turn:    int32(load.Interaction.TurnNumber),
	})
	if err != nil {
		return fmt.Errorf("failed to create messages from interaction: %w", err)
	}
	_, err = s.store.Pg.CreateMessageTx(ctx, tx, database.CreateMessageParams{
		SessionID: pgSessionId,
		UserID:    pgUserId,
		Role:      database.MessagesRoleModel,
		Content:   load.Interaction.ModelOutput.Text,
		Turn:      int32(load.Interaction.TurnNumber),
	})
	if err != nil {
		return fmt.Errorf("failed to create messages from interaction: %w", err)
	}

	return nil
}
