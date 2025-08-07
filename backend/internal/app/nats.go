package app

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	natsConnNameMain              string        = "shaikh-main"
	natsConnTimeoutTenSeconds     time.Duration = 10 * time.Second
	natsPingIntervalTwentySeconds time.Duration = 20 * time.Second
	natsMaxPingsOutstandingFive   int           = 5
	natsReconnectWaitTenSeconds   time.Duration = 10 * time.Second
	jsStreamNameContext           string        = "CONTEXT"
)

func NewNats(opts *nats.Options) (*nats.Conn, error) {
	nc, err := nats.Connect(
		opts.Url,
		nats.Name(opts.Name),
		nats.Timeout(opts.Timeout),
		nats.PingInterval(opts.PingInterval),
		nats.MaxPingsOutstanding(opts.MaxPingsOut),
		nats.ReconnectWait(opts.ReconnectWait),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create nats conn: %w", err)
	}

	return nc, nil
}

func NewJetStream(nc *nats.Conn) (jetstream.JetStream, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create new jetstream: %w", err)
	}

	return js, err
}
