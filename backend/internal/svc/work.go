package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
	"github.com/google/uuid"
)

const (
	SyncerDurableName   string        = "syncer"
	SyncerSubject       string        = ContextStreamSubject + "sync"
	SyncerCommitSubject string        = SyncerSubject + ".commit"
	SyncMaxIdleTime     time.Duration = 2 * time.Minute
	SyncerAckWait       time.Duration = 4 * time.Minute
	SyncMaxBatchSize    int           = 5
)

type SyncPayload struct {
	UserID      uuid.UUID        `json:"user_id"`
	SessionID   uuid.UUID        `json:"session_id"`
	Interaction *dom.Interaction `json:"interaction"`
}

type SyncCommitPayload struct {
	UserID    uuid.UUID `json:"user_id"`
	SessionID uuid.UUID `json:"session_id"`
}

type Syncer struct {
	Cons        dom.PubSubConsumer
	Publisher   dom.Publisher
	UnitOfWork  dom.UnitOfWork
	SessionRepo dom.SessionRepo
	Logger      *slog.Logger
	Buffer      []dom.PubMsg
}

func BuildSyncer(
	ctx context.Context,
	ps dom.PubSub,
	pub dom.Publisher,
	uow dom.UnitOfWork,
	sr dom.SessionRepo,
) (*Syncer, error) {
	log := slog.Default().With(
		"worker", "syncer",
	)

	cons, err := ps.CreateConsumer(ctx, ContextStream, dom.PubSubConsumerConfig{
		Name:           SyncerDurableName,
		Durable:        true,
		AckWait:        SyncerAckWait,
		FilterSubjects: []string{SyncerSubject},
	})
	if err != nil {
		return nil, err
	}

	return &Syncer{
		Cons:        cons,
		Publisher:   pub,
		UnitOfWork:  uow,
		SessionRepo: sr,
		Logger:      log,
		Buffer:      []dom.PubMsg{},
	}, nil
}

func (s *Syncer) Consumer() dom.PubSubConsumer {
	return s.Cons
}

func (s *Syncer) Start(ctx context.Context) error {
	msgs, err := s.Cons.Messages(ctx)
	if err != nil {
		return fmt.Errorf("syncer failed: %w", err)
	}

	timer := time.NewTimer(SyncMaxIdleTime)
	reset := func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(SyncMaxIdleTime)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-timer.C:
			if err := s.IdleProcess(ctx); err != nil {
				return fmt.Errorf("syncer failed: %w", err)
			}
			reset()

		case m, ok := <-msgs:
			if !ok {
				return nil
			}
			if err := s.Process(ctx, m); err != nil {
				return fmt.Errorf("syncer failed: %w", err)
			}
			reset()
		}
	}
}

func (s *Syncer) Process(ctx context.Context, msg dom.PubMsg) error {
	s.Buffer = append(s.Buffer, msg)
	if len(s.Buffer) >= SyncMaxBatchSize {
		if err := s.flush(ctx); err != nil {
			return err
		}
	} else {
		for _, m := range s.Buffer {
			if err := m.InProgress(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Syncer) IdleProcess(ctx context.Context) error {
	if len(s.Buffer) > 0 {
		if err := s.flush(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *Syncer) flush(ctx context.Context) error {
	tx, err := s.UnitOfWork.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	type agg struct {
		MaxTurn int32
	}
	byUS := map[[2]uuid.UUID]*agg{}

	loads := make([]SyncPayload, len(s.Buffer))
	for i, m := range s.Buffer {
		if err := json.Unmarshal(m.Data(), &loads[i]); err != nil {
			m.Term()
			return err
		}

		if err := s.sync(ctx, tx, loads[i]); err != nil {
			return err
		}

		key := [2]uuid.UUID{loads[i].UserID, loads[i].SessionID}
		a := byUS[key]
		if a == nil {
			a = &agg{}
			byUS[key] = a
		}

		t := int32(loads[i].Interaction.TurnNumber)
		if t > a.MaxTurn {
			a.MaxTurn = t
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	for _, m := range s.Buffer {
		if err := m.Ack(); err != nil {
			return err
		}
	}

	s.Buffer = s.Buffer[:0]

	for k, a := range byUS {
		_, err := s.SessionRepo.UpdateSessionByID(ctx, dom.Session{
			ID:      k[1],
			MaxTurn: &a.MaxTurn,
		})
		if err != nil {
			return err
		}
		evt := SyncCommitPayload{
			UserID:    k[0],
			SessionID: k[1],
		}

		b, err := json.Marshal(evt)
		if err != nil {
			return err
		}

		ack, err := s.Publisher.Publish(ctx, SyncerCommitSubject, b, &dom.PubOptions{
			MsgID: fmt.Sprintf("commit:%s:%s:%d", k[0], k[1], a.MaxTurn),
		})

		if ack.Stream != ContextStream {
			return fmt.Errorf("published to unexpected stream: %s", ack.Stream)
		}
	}
	return nil
}

func (s *Syncer) sync(
	ctx context.Context,
	tx dom.Tx,
	load SyncPayload,
) error {
	var mr dom.MessageRepo
	if err := tx.Get(&mr); err != nil {
		return err
	}

	turn := load.Interaction.TurnNumber
	inf1 := load.Interaction.Inferences[0]
	_, err := mr.CreateMessage(ctx, &dom.UserMessage{
		MsgMeta: dom.MsgMeta{
			SessionID:        load.SessionID,
			UserID:           load.UserID,
			Model:            &inf1.Model,
			Turn:             turn,
			TotalInputTokens: &inf1.InputTokens,
			Content:          &inf1.Input.Text,
		},
		MsgContent: inf1.Input.Text,
	})
	if err != nil {
		return err
	}

	if len(load.Interaction.Inferences) > 1 {
		inf2 := load.Interaction.Inferences[1]
		call, err := dom.ToJsonRawMessage(inf2.Output.FunctionCall.Args)
		if err != nil {
			return err
		}
		resp, err := dom.ToJsonRawMessage(inf1.Input.FunctionResponse.Content)
		if err != nil {
			return err
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
			return err
		}

		_, err = mr.CreateMessage(ctx, &dom.ModelMessage{
			MsgMeta: dom.MsgMeta{
				SessionID:         load.SessionID,
				UserID:            load.UserID,
				Model:             &inf2.Model,
				Turn:              turn,
				TotalOutputTokens: &inf2.OutputTokens,
				Content:           &inf2.Output.Text,
			},
			MsgContent: inf2.Output.Text,
		})
		if err != nil {
			return err
		}

		return nil
	}

	_, err = mr.CreateMessage(ctx, &dom.ModelMessage{
		MsgMeta: dom.MsgMeta{
			SessionID:         load.SessionID,
			UserID:            load.UserID,
			Model:             &inf1.Model,
			Turn:              turn,
			TotalOutputTokens: &inf1.OutputTokens,
			Content:           &inf1.Output.Text,
		},
		MsgContent: inf1.Output.Text,
	})
	return nil
}

type Summarizer struct {
	Cons        dom.PubSubConsumer
	Agent       dom.Agent
	SessionRepo dom.SessionRepo
	MessageRepo dom.MessageRepo
	Logger      *slog.Logger
}

const (
	SummarizerDurableName string        = "summarizer"
	SummarizerMaxIdleTime time.Duration = 5 * time.Minute
	SummarizerAckWait     time.Duration = 7 * time.Minute
	SummarizerMinTurns    int           = 10
)

func BuildSummarizer(
	ctx context.Context,
	ps dom.PubSub,
	ag dom.Agent,
	sr dom.SessionRepo,
	mr dom.MessageRepo,
) (*Summarizer, error) {
	log := slog.Default().With("worker", "summarizer")

	cons, err := ps.CreateConsumer(ctx, ContextStream, dom.PubSubConsumerConfig{
		Name:           SummarizerDurableName,
		Durable:        true,
		AckWait:        SummarizerAckWait,
		FilterSubjects: []string{SyncerCommitSubject},
	})
	if err != nil {
		return nil, err
	}

	return &Summarizer{
		Cons:        cons,
		Agent:       ag,
		SessionRepo: sr,
		MessageRepo: mr,
		Logger:      log,
	}, nil
}

func (s *Summarizer) Consumer() dom.PubSubConsumer {
	return s.Cons
}

func (s *Summarizer) Start(ctx context.Context) error {
	msgs, err := s.Cons.Messages(ctx)
	if err != nil {
		return fmt.Errorf("summarizer failed: %w", err)
	}

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			if err := s.IdleProcess(ctx); err != nil {
				return fmt.Errorf("summarizer failed: %w", err)
			}

		case m, ok := <-msgs:
			if !ok {
				return nil
			}
			if err := s.Process(ctx, m); err != nil {
				return fmt.Errorf("summarizer failed: %w", err)
			}
		}
	}
}

func (s *Summarizer) Process(ctx context.Context, msg dom.PubMsg) error {
	var load SyncCommitPayload
	if err := json.Unmarshal(msg.Data(), &load); err != nil {
		_ = msg.Term()
		return err
	}

	sess, err := s.SessionRepo.GetSessionByID(ctx, load.SessionID)
	if err != nil {
		return err
	}

	delta := *sess.MaxTurn - *sess.MaxTurnSummarized
	if delta > int32(SummarizerMinTurns) {
		if err := msg.InProgress(); err != nil {
			return err
		}
		if err := s.summarize(ctx, sess); err != nil {
			return err
		}
	}

	return msg.Ack()
}

func (s *Summarizer) IdleProcess(ctx context.Context) error {
	sessions, err := s.SessionRepo.ListWithBacklog(ctx)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, sess := range sessions {
		if now.Sub(sess.LastAccessed) >= SummarizerMaxIdleTime {
			if err := s.summarize(ctx, sess); err != nil {
				return err
			}
		}
	}
	return nil
}

type SummarizerResponse struct {
	Summary string `json:"summary"`
}

func (s *Summarizer) summarize(
	ctx context.Context,
	sess dom.Session,
) error {
	msgs, err := s.MessageRepo.GetMessagesBySessionIDOrdered(ctx, sess.ID)
	if err != nil {
		return err
	}
	if len(msgs) == 0 {
		return nil
	}

	lastTurn := msgs[len(msgs)-1].Meta().Turn

	win, err := dom.MessagesToLLMContent(msgs)
	if err != nil {
		return err
	}

	res, err := s.Agent.Generate(ctx, dom.Summarizer, win)
	if err != nil {
		return err
	}
	if res.Bytes == nil {
		return fmt.Errorf("empty summarizer result")
	}

	var resp SummarizerResponse
	if err := json.Unmarshal(res.Bytes, &resp); err != nil {
		return err
	}

	_, err = s.SessionRepo.UpdateSessionByID(ctx, dom.Session{
		ID:                sess.ID,
		LastAccessed:      time.Now(),
		MaxTurn:           &lastTurn,
		MaxTurnSummarized: &lastTurn,
		Summary:           &resp.Summary,
	})
	return err
}

const (
	MemorizerDurableName string = "memorizer"
)

type Memorizer struct {
	Cons        dom.PubSubConsumer
	Agent       dom.Agent
	MemoryRepo  dom.MemoryRepo
	MessageRepo dom.MessageRepo
	Logger      *slog.Logger
}

func BuildMemorizer(
	ctx context.Context,
	ps dom.PubSub,
	ag dom.Agent,
	mr dom.MessageRepo,
	memr dom.MemoryRepo,
) (*Memorizer, error) {
	log := slog.Default().With(
		"worker", "memorizer",
	)

	cons, err := ps.CreateConsumer(ctx, ContextStream, dom.PubSubConsumerConfig{
		Name:           MemorizerDurableName,
		Durable:        true,
		FilterSubjects: []string{SyncerCommitSubject},
	})
	if err != nil {
		return nil, err
	}

	return &Memorizer{
		Cons:        cons,
		Agent:       ag,
		MemoryRepo:  memr,
		MessageRepo: mr,
		Logger:      log,
	}, nil
}

func (m *Memorizer) Consumer() dom.PubSubConsumer {
	return m.Cons
}

func (m *Memorizer) Start(ctx context.Context) error {
	msgs, err := m.Cons.Messages(ctx)
	if err != nil {
		return fmt.Errorf("memorizer failed: %w", err)
	}

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			if err := s.IdleProcess(ctx); err != nil {
				return fmt.Errorf("summarizer failed: %w", err)
			}

		case m, ok := <-msgs:
			if !ok {
				return nil
			}
			if err := s.Process(ctx, m); err != nil {
				return fmt.Errorf("summarizer failed: %w", err)
			}
		}
	}
}

func (m *Memorizer) Process(ctx context.Context, msg dom.PubMsg) error {
	var load SyncCommitPayload
	if err := json.Unmarshal(msg.Data(), &load); err != nil {
		_ = msg.Term()
		return err
	}

	sess, err := s.SessionRepo.GetSessionByID(ctx, load.SessionID)
	if err != nil {
		return err
	}

	delta := *sess.MaxTurn - *sess.MaxTurnSummarized
	if delta > int32(SummarizerMinTurns) {
		if err := msg.InProgress(); err != nil {
			return err
		}
		if err := s.summarize(ctx, sess); err != nil {
			return err
		}
	}

	return msg.Ack()
}
