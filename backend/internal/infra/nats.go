package infra

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

type NatsClient struct {
	Js  jetstream.JetStream
	Log *slog.Logger
}

func (c *NatsClient) Publish(
	ctx context.Context,
	subject string,
	data []byte,
	opts dom.PubOptions,
) (*dom.PubAck, error) {
	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
	}

	msg.Header = nats.Header{}
	msg.Header.Set("Nats-Msg-Id", opts.MsgID)

	ack, err := c.Js.PublishMsg(ctx, msg)
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
