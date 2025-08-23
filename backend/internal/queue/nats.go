package queue

import (
	"time"
)

const (
	natsConnNameApi               string        = "shaikh-api"
	natsConnNameWorker            string        = "shaikh-worker"
	natsConnTimeoutTenSeconds     time.Duration = 10 * time.Second
	natsPingIntervalTwentySeconds time.Duration = 20 * time.Second
	natsMaxPingsOutstandingFive   int           = 5
	natsReconnectWaitTenSeconds   time.Duration = 10 * time.Second
)