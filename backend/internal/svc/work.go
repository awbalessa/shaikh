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
	SyncerAckWait       time.Duration = 4 * time.Minute
	SyncMaxIdleTime     time.Duration = 2 * time.Minute
	SyncMaxBatchSize    int           = 5
)

type SyncPayload struct {
	UserID      uuid.UUID        `json:"user_id"`
	SessionID   uuid.UUID        `json:"session_id"`
	Interaction *dom.Interaction `json:"interaction"`
}

type Syncer struct {
	Cons       dom.PubSubConsumer
	Publisher  dom.Publisher
	UnitOfWork dom.UnitOfWork
	Logger     *slog.Logger
	Buffer     []dom.PubMsg
}

func BuildSyncer(
	ctx context.Context,
	ps dom.PubSub,
	pub dom.Publisher,
	uow dom.UnitOfWork,
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
		Cons:       cons,
		Publisher:  pub,
		UnitOfWork: uow,
		Logger:     log,
		Buffer:     []dom.PubMsg{},
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
		MaxTurn   int32
		DeltaMsgs int32
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
		a.DeltaMsgs++
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
		evt := SyncCommitPayload{
			UserID: k[0], SessionID: k[1],
			MaxTurn:   a.MaxTurn,
			DeltaMsgs: a.DeltaMsgs,
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

type sessionState struct {
	userID             uuid.UUID
	lastTurnSeen       int32
	lastTurnSummarized int32
	pendingMsgs        int32
	lastActivity       time.Time
}

type SyncCommitPayload struct {
	UserID    uuid.UUID `json:"user_id"`
	SessionID uuid.UUID `json:"session_id"`
	MaxTurn   int32     `json:"max_turn"`
	DeltaMsgs int32     `json:"delta_msgs"`
}

type Summarizer struct {
	Cons        dom.PubSubConsumer
	Agent       dom.Agent
	SessionRepo dom.SessionRepo
	MessageRepo dom.MessageRepo
	Logger      *slog.Logger
	Sessions    map[uuid.UUID]*sessionState
}

const (
	SummarizerDurableName  string        = "summarizer"
	SummarizerMaxIdleTime  time.Duration = 5 * time.Minute
	SummarizerMaxBatchSize int           = 5
	SummarizerMinTurns     int           = 10
)

func BuildSummarizer(
	ctx context.Context,
	ps dom.PubSub,
	ag dom.Agent,
	sr dom.SessionRepo,
) (*Summarizer, error) {
	log := slog.Default().With(
		"worker", "summarizer",
	)

	cons, err := ps.CreateConsumer(ctx, ContextStream, dom.PubSubConsumerConfig{
		Name:           SummarizerDurableName,
		Durable:        true,
		FilterSubjects: []string{SyncerCommitSubject},
	})
	if err != nil {
		return nil, err
	}

	sesh := make(map[uuid.UUID]*sessionState)

	return &Summarizer{
		Cons:        cons,
		Agent:       ag,
		SessionRepo: sr,
		Logger:      log,
		Sessions:    sesh,
	}, nil
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
		msg.Term()
		return err
	}

	st, err := s.getState(ctx, load.SessionID, load.UserID, load.MaxTurn)
	if err != nil {
		return err
	}

	st.pendingMsgs += load.DeltaMsgs
	st.lastTurnSeen = load.MaxTurn
	st.lastActivity = time.Now()
	deltaTurns := st.lastTurnSeen - st.lastTurnSummarized
	if deltaTurns > int32(SummarizerMinTurns) {
		if err := s.summarize(ctx, load.SessionID); err != nil {
			return err
		}
		st.lastTurnSummarized = st.lastTurnSeen
		st.pendingMsgs = 0
	}

	return nil
}

func (s *Summarizer) IdleProcess(ctx context.Context) error {
	for sid, st := range s.Sessions {
		if st.lastTurnSeen > st.lastTurnSummarized && time.Since(st.lastActivity) >= SummarizerMaxIdleTime {
			if err := s.summarize(ctx, sid); err != nil {
				return err
			}

			st.lastTurnSummarized = st.lastTurnSeen
			st.pendingMsgs = 0
		}
	}

	return nil
}

func (s *Summarizer) getState(
	ctx context.Context,
	sessionID, userID uuid.UUID,
	lastSeen int32,
) (*sessionState, error) {
	st := s.Sessions[sessionID]
	if st != nil {
		return st, nil
	}

	last, err := s.SessionRepo.GetMaxTurnByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	st = &sessionState{
		userID:             userID,
		lastTurnSeen:       lastSeen,
		lastTurnSummarized: last,
	}
	s.Sessions[sessionID] = st
	return st, nil
}

type SummarizerResponse struct {
	Summary string `json:"summary"`
}

func (s *Summarizer) summarize(
	ctx context.Context,
	sessionID uuid.UUID,
) error {
	msgs, err := s.MessageRepo.GetMessagesBySessionIDOrdered(ctx, sessionID)
	if err != nil {
		return err
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
		return err
	}

	var resp SummarizerResponse
	if err := json.Unmarshal(res.Bytes, &resp); err != nil {
		return err
	}

	_, err = s.SessionRepo.UpdateSessionByID(ctx, dom.Session{
		ID:           sessionID,
		LastAccessed: time.Now(),
		MaxTurn:      &lastTurn,
		Summary:      &resp.Summary,
	})
	if err != nil {
		return err
	}

	return nil
}

type Memorizer struct {
	Cons        dom.PubSubConsumer
	Agent       dom.Agent
	MemoryRepo  dom.MemoryRepo
	MessageRepo dom.MessageRepo
	Logger      *slog.Logger
}

func BuildMemorizer()
