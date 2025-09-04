package pro

import (
	"context"
	"fmt"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Nats struct {
	Conn *nats.Conn
}

func NewNats(name string) (*Nats, error) {
	nc, err := nats.Connect(
		nats.DefaultURL,
		nats.Name(name),
		nats.Timeout(10*time.Second),
		nats.PingInterval(20*time.Second),
		nats.MaxPingsOutstanding(5),
		nats.ReconnectWait(10*time.Second),
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

func NewNatsPubSub(nats *Nats, js jetstream.JetStream) *NatsPubSub {
	return &NatsPubSub{Conn: nats.Conn, Js: js}
}

func (n *NatsPubSub) Publisher() dom.Publisher {
	return &NatsPublisher{
		Conn: n.Conn,
		Js:   n.Js,
	}
}

func (n *NatsPubSub) Subscriber() dom.Subscriber {
	return &NatsSubscriber{
		Conn: n.Conn,
	}
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

type NatsDurablePubMsg struct {
	Msg jetstream.Msg
}

func (m *NatsDurablePubMsg) Data() []byte {
	return m.Msg.Data()
}

func (m *NatsDurablePubMsg) Subject() string {
	return m.Msg.Subject()
}

func (m *NatsDurablePubMsg) Ack() error {
	return m.Msg.Ack()
}

func (m *NatsDurablePubMsg) Nak() error {
	return m.Msg.Nak()
}

func (m *NatsDurablePubMsg) Term() error {
	return m.Msg.Term()
}

func (m *NatsDurablePubMsg) InProgress() error {
	return m.Msg.InProgress()
}

func (m *NatsDurablePubMsg) Metadata() (dom.DurablePubMsgMetadata, error) {
	meta, err := m.Msg.Metadata()
	if err != nil {
		return dom.DurablePubMsgMetadata{}, err
	}

	return dom.DurablePubMsgMetadata{
		Stream:       meta.Stream,
		Consumer:     meta.Consumer,
		NumDelivered: meta.NumDelivered,
		Timestamp:    meta.Timestamp,
	}, nil
}

type NatsPubSubConsumer struct {
	Cons jetstream.Consumer
}

func (c *NatsPubSubConsumer) Fetch(batch int) ([]dom.DurablePubMsg, error) {
	msgs, err := c.Cons.Fetch(batch)
	if err != nil {
		return nil, err
	}

	if msgs.Error() != nil {
		return nil, msgs.Error()
	}

	var final []dom.DurablePubMsg
	for m := range msgs.Messages() {
		final = append(final, toNatsMsg(m))
	}

	return final, nil
}

func (c *NatsPubSubConsumer) Messages(ctx context.Context) (<-chan dom.DurablePubMsg, error) {
	msgs, err := c.Cons.Messages()
	if err != nil {
		return nil, err
	}

	out := make(chan dom.DurablePubMsg)

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
	Conn *nats.Conn
	Js   jetstream.JetStream
}

func (c *NatsPublisher) Publish(subject string, data []byte) error {
	return c.Conn.Publish(subject, data)
}

func (c *NatsPublisher) Request(ctx context.Context, subject string, data []byte) (*dom.PubMsg, error) {
	msg, err := c.Conn.RequestWithContext(ctx, subject, data)
	if err != nil {
		return nil, err
	}

	if msg == nil {
		return nil, fmt.Errorf("nats: nil reply for subject %s", subject)
	}

	return &dom.PubMsg{
		Subject: msg.Subject,
		Reply:   msg.Reply,
		Data:    msg.Data,
	}, nil
}

func (c *NatsPublisher) DurablePublish(
	ctx context.Context,
	subject string,
	data []byte,
	opts *dom.DurablePubOptions,
) (*dom.DurablePubAck, error) {
	ack, err := c.Js.Publish(ctx, subject, data,
		jetstream.WithMsgID(opts.MsgID),
	)
	if err != nil {
		return nil, err
	}

	if ack == nil {
		return nil, fmt.Errorf("unexpected publish ack: %+v", ack)
	}

	return &dom.DurablePubAck{
		Stream: ack.Stream,
		Seq:    ack.Sequence,
	}, nil
}

type NatsSubscriber struct {
	Conn *nats.Conn
}

func (s *NatsSubscriber) Subscribe(subject string, handler func(msg *dom.PubMsg)) error {
	_, err := s.Conn.Subscribe(subject, func(m *nats.Msg) {
		dommsg := &dom.PubMsg{
			Subject: m.Subject,
			Reply:   m.Reply,
			Data:    m.Data,
		}
		handler(dommsg)
	})
	return err
}

var toJsRetentionPolicy = map[dom.PubSubRetentionPolicy]jetstream.RetentionPolicy{
	dom.WorkQueue:   jetstream.WorkQueuePolicy,
	dom.LimitsBased: jetstream.LimitsPolicy,
}

var toJsStorage = map[dom.PubSubStorageType]jetstream.StorageType{
	dom.FileStorage: jetstream.FileStorage,
}

func toNatsMsg(msg jetstream.Msg) dom.DurablePubMsg {
	return &NatsDurablePubMsg{
		Msg: msg,
	}
}
