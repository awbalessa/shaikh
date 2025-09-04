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
	SyncerSubject       string        = ContextStreamSubject + "." + "sync"
	SyncerCommitSubject string        = SyncerSubject + "." + "commit"
	SyncMaxIdleTime     time.Duration = 2 * time.Minute
	AckWaitOffset       time.Duration = 2 * time.Minute
	SyncerAckWait       time.Duration = SyncMaxIdleTime + AckWaitOffset
	SyncMaxBatchSize    int           = 5
	SyncerPingSubject                 = "ping.syncer"
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
	UserRepo    dom.UserRepo
	Logger      *slog.Logger
	Buffer      []dom.DurablePubMsg
}

func BuildSyncer(
	ctx context.Context,
	ps dom.PubSub,
	uow dom.UnitOfWork,
	sr dom.SessionRepo,
	ur dom.UserRepo,
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

	pub := ps.Publisher()

	if err := ps.Subscriber().Subscribe(SyncerPingSubject, func(msg *dom.PubMsg) {
		if msg.Reply != "" {
			resp := []byte(`{"status": "ok"}`)
			pub.Publish(msg.Reply, resp)
		}

	}); err != nil {
		return nil, err
	}

	return &Syncer{
		Cons:        cons,
		Publisher:   pub,
		UnitOfWork:  uow,
		SessionRepo: sr,
		UserRepo:    ur,
		Logger:      log,
		Buffer:      []dom.DurablePubMsg{},
	}, nil
}

func (s *Syncer) Consumer() dom.PubSubConsumer {
	return s.Cons
}

func (s *Syncer) Start(ctx context.Context) error {
	msgs, _, err := s.Cons.Messages(ctx)
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

func (s *Syncer) Process(ctx context.Context, msg dom.DurablePubMsg) error {
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

	type sessAgg struct {
		MaxTurn int32
	}
	type userAgg struct {
		DeltaMsgs int32
	}

	bySession := map[uuid.UUID]*sessAgg{}
	byUser := map[uuid.UUID]*userAgg{}

	loads := make([]SyncPayload, len(s.Buffer))
	for i, m := range s.Buffer {
		if err := json.Unmarshal(m.Data(), &loads[i]); err != nil {
			m.Term()
			return err
		}
		if err := s.sync(ctx, tx, loads[i]); err != nil {
			return err
		}

		if sa, ok := bySession[loads[i].SessionID]; ok {
			if loads[i].Interaction.TurnNumber > sa.MaxTurn {
				sa.MaxTurn = loads[i].Interaction.TurnNumber
			}
		} else {
			bySession[loads[i].SessionID] = &sessAgg{
				MaxTurn: loads[i].Interaction.TurnNumber,
			}
		}

		if ua, ok := byUser[loads[i].UserID]; ok {
			ua.DeltaMsgs++
		} else {
			byUser[loads[i].UserID] = &userAgg{DeltaMsgs: 1}
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

	for sid, sa := range bySession {
		if _, err := s.SessionRepo.UpdateSessionByID(ctx, sid, &sa.MaxTurn, nil, nil, nil); err != nil {
			return err
		}
	}

	for uid, ua := range byUser {
		if _, err := s.UserRepo.IncrementUserMessagesByID(ctx, uid, ua.DeltaMsgs, 0); err != nil {
			return err
		}
	}

	for sid := range bySession {
		evt := SyncCommitPayload{
			UserID:    findUserForSession(loads, sid),
			SessionID: sid,
		}
		b, err := json.Marshal(evt)
		if err != nil {
			return err
		}
		ack, err := s.Publisher.DurablePublish(ctx, SyncerCommitSubject, b, &dom.DurablePubOptions{
			MsgID: fmt.Sprintf("sync.commit:%s:%s:%d", evt.UserID, evt.SessionID, bySession[evt.SessionID].MaxTurn),
		})
		if err != nil {
			return err
		}
		if ack.Stream != ContextStream {
			return fmt.Errorf("published to unexpected stream: %s", ack.Stream)
		}
	}

	return nil
}

func findUserForSession(loads []SyncPayload, sid uuid.UUID) uuid.UUID {
	for _, l := range loads {
		if l.SessionID == sid {
			return l.UserID
		}
	}
	return uuid.Nil
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
	SummarizerAckWait     time.Duration = SummarizerMaxIdleTime + AckWaitOffset
	SummarizerMinTurns    int32         = 10
	SummarizerPingSubject               = "ping.summarizer"
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

	if err := ps.Subscriber().Subscribe(SummarizerPingSubject, func(msg *dom.PubMsg) {
		if msg.Reply != "" {
			resp := []byte(`{"status": "ok"}`)
			ps.Publisher().Publish(msg.Reply, resp)
		}

	}); err != nil {
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
	msgs, _, err := s.Cons.Messages(ctx)
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

func (s *Summarizer) Process(ctx context.Context, msg dom.DurablePubMsg) error {
	var load SyncCommitPayload
	if err := json.Unmarshal(msg.Data(), &load); err != nil {
		_ = msg.Term()
		return err
	}

	sess, err := s.SessionRepo.GetSessionByID(ctx, load.SessionID)
	if err != nil {
		return err
	}

	delta := sess.MaxTurn - sess.MaxTurnSummarized
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
	sessions, err := s.SessionRepo.ListSessionsWithBacklog(ctx)
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
	sess *dom.Session,
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

	_, err = s.SessionRepo.UpdateSessionByID(
		ctx,
		sess.ID,
		&lastTurn,
		&lastTurn,
		&resp.Summary,
		nil,
	)
	return err
}

const (
	MemorizerDurableName string        = "memorizer"
	MemorizerMinMsgs     int32         = 50
	MemorizerMinMsgsIdle int32         = 10
	MemorizerMaxIdleTime time.Duration = 30 * time.Minute
	MemorizerAckWait     time.Duration = MemorizerMaxIdleTime + AckWaitOffset
	MemorizerPingSubject string        = "ping.memorizer"
)

type Memorizer struct {
	Cons        dom.PubSubConsumer
	Agent       dom.Agent
	MemoryRepo  dom.MemoryRepo
	MessageRepo dom.MessageRepo
	UserRepo    dom.UserRepo
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

	if err := ps.Subscriber().Subscribe(MemorizerPingSubject, func(msg *dom.PubMsg) {
		if msg.Reply != "" {
			resp := []byte(`{"status": "ok"}`)
			ps.Publisher().Publish(msg.Reply, resp)
		}

	}); err != nil {
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
	msgs, _, err := m.Cons.Messages(ctx)
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
			if err := m.IdleProcess(ctx); err != nil {
				return fmt.Errorf("memorizer failed: %w", err)
			}

		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			if err := m.Process(ctx, msg); err != nil {
				return fmt.Errorf("memorizer failed: %w", err)
			}
		}
	}
}

func (m *Memorizer) Process(ctx context.Context, msg dom.DurablePubMsg) error {
	var load SyncCommitPayload
	if err := json.Unmarshal(msg.Data(), &load); err != nil {
		_ = msg.Term()
		return err
	}

	user, err := m.UserRepo.GetUserByID(ctx, load.UserID)
	if err != nil {
		return err
	}

	delta := user.TotalMessages - user.TotalMessagesMemorized
	if delta > MemorizerMinMsgs {
		if err := msg.InProgress(); err != nil {
			return err
		}
		if err := m.memorize(ctx, user); err != nil {
			return err
		}
	}

	return msg.Ack()
}

func (m *Memorizer) IdleProcess(ctx context.Context) error {
	users, err := m.UserRepo.ListUsersWithBacklog(ctx)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, u := range users {
		if now.Sub(u.UpdatedAt) >= MemorizerMaxIdleTime {
			if err := m.memorize(ctx, u); err != nil {
				return err
			}
		}
	}
	return nil
}

type MemorizerMemory struct {
	UniqueKey  string  `json:"unique_key"`
	Content    string  `json:"content"`
	Confidence float32 `json:"confidence"`
	SourceMsg  string  `json:"source_msg"`
}

type MemorizerResponse struct {
	Memories   []MemorizerMemory `json:"memories"`
	DeleteKeys []string          `json:"delete_keys"`
}

func (m *Memorizer) memorize(
	ctx context.Context,
	user *dom.User,
) error {
	mems, err := m.MemoryRepo.GetMemoriesByUserID(ctx, user.ID, 50)
	if err != nil {
		return err
	}

	memwin, err := dom.MemoriesToLLMContent(mems)
	if err != nil {
		return err
	}

	msgs, err := m.MessageRepo.GetUserMessagesByUserID(ctx, user.ID, 100)
	if err != nil {
		return err
	}
	if len(msgs) == 0 {
		return nil
	}

	msgwin, err := dom.MessagesToLLMContent(msgs)
	if err != nil {
		return err
	}

	full := append(memwin, msgwin...)

	res, err := m.Agent.Generate(ctx, dom.Memorizer, full)
	if err != nil {
		return err
	}
	if res.Bytes == nil {
		return fmt.Errorf("empty memorizer result")
	}

	var mr MemorizerResponse
	if err := json.Unmarshal(res.Bytes, &mr); err != nil {
		return err
	}

	for _, key := range mr.DeleteKeys {
		if err := m.MemoryRepo.DeleteMemoryByUserIDKey(ctx, user.ID, key); err != nil {
			return err
		}
	}

	for _, mem := range mr.Memories {
		if mem.Confidence > 0.75 {
			_, err := m.MemoryRepo.UpsertMemory(
				ctx,
				user.ID,
				mem.SourceMsg,
				mem.Confidence,
				mem.UniqueKey,
				mem.Content,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type WorkerProbe struct {
	name string
	subj string
	pub  dom.Publisher
}

func NewWorkerProbe(worker string, subject string, ps dom.PubSub) *WorkerProbe {
	return &WorkerProbe{
		name: worker,
		subj: subject,
		pub:  ps.Publisher(),
	}
}

func (p *WorkerProbe) Name() string {
	return p.name
}

func (p *WorkerProbe) Ping(ctx context.Context) error {
	msg, err := p.pub.Request(ctx, p.subj, nil)
	if err != nil {
		return err
	}
	var resp struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		return err
	}

	if resp.Status != "ok" {
		return fmt.Errorf("worker %s down", p.name)
	}

	return nil
}
