package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
)

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

	turn := load.InteractionDTO.TurnNumber
	inf1 := load.InteractionDTO.Inferences[0]
	_, err := mr.CreateMessage(ctx, &dom.UserMessage{
		MsgMeta: dom.MsgMeta{
			SessionID:        load.SessionID,
			UserID:           load.UserID,
			Model:            inf1.Model,
			Turn:             turn,
			TotalInputTokens: &inf1.InputTokens,
			Content:          &inf1.Input.Text,
		},
		MsgContent: inf1.Input.Text,
	})
	if err != nil {
		return fmt.Errorf("failed to persist messages: %w", err)
	}

	if len(load.InteractionDTO.Inferences) > 1 {
		inf2 := load.InteractionDTO.Inferences[1]
		call, err := toJsonRawMessage(inf2.Output.FunctionCall.Args)
		if err != nil {
			return fmt.Errorf("failed to persist messages: %w", err)
		}
		resp, err := toJsonRawMessage(inf1.Input.FunctionResponse.Content)
		if err != nil {
			return fmt.Errorf("failed to persist messages: %w", err)
		}
		_, err = mr.CreateMessage(ctx, &dom.FunctionMessage{
			MsgMeta: dom.MsgMeta{
				SessionID:         load.SessionID,
				UserID:            load.UserID,
				Turn:              turn,
				TotalInputTokens:  &inf2.InputTokens,
				TotalOutputTokens: &inf1.OutputTokens,
				FunctionName:      &inf1.Input.FunctionResponse.Name,
				FunctionCall:      call,
				FunctionResponse:  resp,
			},
			FunctionName:     inf1.Input.FunctionResponse.Name,
			FunctionCall:     call,
			FunctionResponse: resp,
		})
		if err != nil {
			return fmt.Errorf("failed to persist messages: %w", err)
		}

		_, err = mr.CreateMessage(ctx, &dom.ModelMessage{
			MsgMeta: dom.MsgMeta{
				SessionID:         load.SessionID,
				UserID:            load.UserID,
				Model:             inf2.Model,
				Turn:              turn,
				TotalOutputTokens: &inf2.OutputTokens,
				Content:           &inf2.Output.Text,
			},
			MsgContent: inf2.Output.Text,
		})
		if err != nil {
			return fmt.Errorf("failed to persist messages: %w", err)
		}

		return nil
	}

	_, err = mr.CreateMessage(ctx, &dom.ModelMessage{
		MsgMeta: dom.MsgMeta{
			SessionID:         load.SessionID,
			UserID:            load.UserID,
			Model:             inf1.Model,
			Turn:              turn,
			TotalOutputTokens: &inf1.OutputTokens,
			Content:           &inf1.Output.Text,
		},
		MsgContent: inf1.Output.Text,
	})
	return nil
}
