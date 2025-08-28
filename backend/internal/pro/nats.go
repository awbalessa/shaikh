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
	NatsConnNameWorker            string        = "shaikh-worker"
	natsConnTimeoutTenSeconds     time.Duration = 10 * time.Second
	natsPingIntervalTwentySeconds time.Duration = 20 * time.Second
	natsMaxPingsOutstandingFive   int           = 5
	natsReconnectWaitTenSeconds   time.Duration = 10 * time.Second
)

type Nats struct {
	Conn *nats.Conn
	Log  *slog.Logger
}

func NewNats(name string, log *slog.Logger) (*Nats, error) {
	nc, err := nats.Connect(
		nats.DefaultURL,
		nats.Name(name),
		nats.Timeout(natsConnTimeoutTenSeconds),
		nats.PingInterval(natsPingIntervalTwentySeconds),
		nats.MaxPingsOutstanding(natsMaxPingsOutstandingFive),
		nats.ReconnectWait(natsReconnectWaitTenSeconds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create nats: %w", err)
	}

	return &Nats{
		Conn: nc,
		Log:  log,
	}, nil
}

func NewJS(nats *Nats) (jetstream.JetStream, error) {
	js, err := jetstream.New(nats.Conn)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream client: %w", err)
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
		return fmt.Errorf("failed to create stream: %w", err)
	}

	return nil
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
		c.Log.With(
			"err", err,
		).ErrorContext(ctx, "failed to publish message")
		return nil, fmt.Errorf("failed to publish message: %w", err)
	}

	if ack == nil {
		c.Log.With(
			"ack", ack,
		).ErrorContext(ctx, "unexpected publish ack")
		return nil, fmt.Errorf("unexpected publish ack: %+v", ack)
	}

	return &dom.PubAck{
		Stream: ack.Stream,
		Seq:    ack.Sequence,
	}, nil
}
