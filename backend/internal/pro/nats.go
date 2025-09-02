package pro

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	NatsConnNameApi               string        = "shaikh-api"
	NatsConnNameWorkers           string        = "shaikh-workers"
	natsConnTimeoutTenSeconds     time.Duration = 10 * time.Second
	natsPingIntervalTwentySeconds time.Duration = 20 * time.Second
	natsMaxPingsOutstandingFive   int           = 5
	natsReconnectWaitTenSeconds   time.Duration = 10 * time.Second
)

type Nats struct {
	Conn *nats.Conn
	Log  *slog.Logger
}

func NewNats(name string) (*Nats, error) {
	nc, err := nats.Connect(
		nats.DefaultURL,
		nats.Name(name),
		nats.Timeout(natsConnTimeoutTenSeconds),
		nats.PingInterval(natsPingIntervalTwentySeconds),
		nats.MaxPingsOutstanding(natsMaxPingsOutstandingFive),
		nats.ReconnectWait(natsReconnectWaitTenSeconds),
	)
	if err != nil {
		return nil, err
	}

	return &Nats{
		Conn: nc,
	}, nil
}

func (n *Nats) Ping(ctx context.Context) error {
	if !n.Conn.IsConnected() {
		return fmt.Errorf("nats ping failed: not connected")
	}

	if err := n.Conn.FlushWithContext(ctx); err != nil {
		return fmt.Errorf("nats ping failed: %w", err)
	}

	return nil
}

func (n *Nats) Name() string {
	return "pubsub"
}

func NewJS(nats *Nats) (jetstream.JetStream, error) {
	js, err := jetstream.New(nats.Conn)
	if err != nil {
		return nil, err
	}

	return js, nil
}

type NatsPubSub struct {
	Conn *nats.Conn
	Js   jetstream.JetStream
}

var toJsRetentionPolicy = map[dom.PubSubRetentionPolicy]jetstream.RetentionPolicy{
	dom.WorkQueue:   jetstream.WorkQueuePolicy,
	dom.LimitsBased: jetstream.LimitsPolicy,
}

var toJsStorage = map[dom.PubSubStorageType]jetstream.StorageType{
	dom.FileStorage: jetstream.FileStorage,
}

func (n *NatsPubSub) CreateStream(ctx context.Context, cfg dom.PubSubStreamConfig) error {
	_, err := n.Js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:       cfg.Name,
		Subjects:   cfg.Subjects,
		Retention:  toJsRetentionPolicy[cfg.Retention],
		MaxMsgs:    cfg.MaxMsgs,
		MaxAge:     cfg.MaxAge,
		Storage:    toJsStorage[cfg.Storage],
		Replicas:   cfg.Replicas,
		Duplicates: cfg.Duplicates,
	})
	if err != nil {
		return err
	}

	return nil
}

type NatsPubMsg struct {
	Msg jetstream.Msg
}

func (m *NatsPubMsg) Data() []byte {
	return m.Msg.Data()
}

func (m *NatsPubMsg) Subject() string {
	return m.Msg.Subject()
}

func (m *NatsPubMsg) Ack() error {
	return m.Msg.Ack()
}

func (m *NatsPubMsg) Nak() error {
	return m.Msg.Nak()
}

func (m *NatsPubMsg) Term() error {
	return m.Msg.Term()
}

func (m *NatsPubMsg) InProgress() error {
	return m.Msg.InProgress()
}

func (m *NatsPubMsg) Metadata() (dom.PubMsgMetadata, error) {
	meta, err := m.Msg.Metadata()
	if err != nil {
		return dom.PubMsgMetadata{}, err
	}

	return dom.PubMsgMetadata{
		Stream:       meta.Stream,
		Consumer:     meta.Consumer,
		NumDelivered: meta.NumDelivered,
		Timestamp:    meta.Timestamp,
	}, nil
}

type NatsPubSubConsumer struct {
	Cons jetstream.Consumer
}

func toNatsMsg(msg jetstream.Msg) dom.PubMsg {
	return &NatsPubMsg{
		Msg: msg,
	}
}

func (c *NatsPubSubConsumer) Fetch(batch int) ([]dom.PubMsg, error) {
	msgs, err := c.Cons.Fetch(batch)
	if err != nil {
		return nil, err
	}

	if msgs.Error() != nil {
		return nil, msgs.Error()
	}

	var final []dom.PubMsg
	for m := range msgs.Messages() {
		final = append(final, toNatsMsg(m))
	}

	return final, nil
}

func (c *NatsPubSubConsumer) Messages(ctx context.Context) (<-chan dom.PubMsg, error) {
	msgs, err := c.Cons.Messages()
	if err != nil {
		return nil, err
	}

	out := make(chan dom.PubMsg)

	go func() {
		defer close(out)
		defer msgs.Drain()

		for {
			m, err := msgs.Next()
			if err != nil {
				return
			}

			select {
			case <-ctx.Done():
				return
			case out <- toNatsMsg(m):
			}
		}
	}()

	return out, nil
}

func (n *NatsPubSub) CreateConsumer(
	ctx context.Context,
	stream string,
	cfg dom.PubSubConsumerConfig,
) (dom.PubSubConsumer, error) {
	var cons jetstream.Consumer
	var err error
	if cfg.Durable {
		cons, err = n.Js.CreateOrUpdateConsumer(ctx, stream, jetstream.ConsumerConfig{
			Durable:           cfg.Name,
			InactiveThreshold: cfg.InactiveThreshold,
			DeliverPolicy:     jetstream.DeliverPolicy(cfg.DeliverPolicy),
			AckPolicy:         jetstream.AckPolicy(cfg.AckPolicy),
			AckWait:           cfg.AckWait,
			MaxDeliver:        cfg.MaxDeliver,
			BackOff:           cfg.BackOff,
			FilterSubjects:    cfg.FilterSubjects,
			ReplayPolicy:      jetstream.ReplayPolicy(cfg.ReplayPolicy),
			MaxRequestBatch:   cfg.MaxRequestBatch,
			MaxRequestExpires: cfg.MaxRequestExpires,
		})
		if err != nil {
			return nil, err
		}

	} else {
		cons, err = n.Js.CreateOrUpdateConsumer(ctx, stream, jetstream.ConsumerConfig{
			Name:              cfg.Name,
			InactiveThreshold: cfg.InactiveThreshold,
			DeliverPolicy:     jetstream.DeliverPolicy(cfg.DeliverPolicy),
			AckPolicy:         jetstream.AckPolicy(cfg.AckPolicy),
			AckWait:           cfg.AckWait,
			MaxDeliver:        cfg.MaxDeliver,
			BackOff:           cfg.BackOff,
			FilterSubjects:    cfg.FilterSubjects,
			ReplayPolicy:      jetstream.ReplayPolicy(cfg.ReplayPolicy),
			MaxRequestBatch:   cfg.MaxRequestBatch,
			MaxRequestExpires: cfg.MaxRequestExpires,
		})
		if err != nil {
			return nil, err
		}
	}

	return &NatsPubSubConsumer{
		Cons: cons,
	}, nil
}

type NatsPublisher struct {
	Js  jetstream.JetStream
	Log *slog.Logger
}

func NewNatsPublisher(js jetstream.JetStream, log *slog.Logger) *NatsPublisher {
	return &NatsPublisher{Js: js, Log: log}
}

func (c *NatsPublisher) Publish(
	ctx context.Context,
	subject string,
	data []byte,
	opts dom.PubOptions,
) (*dom.PubAck, error) {
	ack, err := c.Js.Publish(ctx, subject, data,
		jetstream.WithMsgID(opts.MsgID),
	)
	if err != nil {
		return nil, err
	}

	if ack == nil {
		return nil, fmt.Errorf("unexpected publish ack: %+v", ack)
	}

	return &dom.PubAck{
		Stream: ack.Stream,
		Seq:    ack.Sequence,
	}, nil
}
