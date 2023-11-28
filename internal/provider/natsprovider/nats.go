package natsprovider

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

const (
	defaultConnectTimeout = 3 * time.Second
)

type Config struct {
	StanClusterID  string
	ClientID       string
	URL            string
	ConnectTimeout time.Duration
}

type NatsProvider struct {
	stan.Conn
}

func New(cfg Config) (*NatsProvider, error) {
	connectTimeout := defaultConnectTimeout
	if cfg.ConnectTimeout != 0 {
		connectTimeout = cfg.ConnectTimeout
	}

	sc, err := stan.Connect(cfg.StanClusterID, cfg.ClientID, stan.NatsURL(cfg.URL), stan.NatsOptions(
		nats.Timeout(connectTimeout),
	))
	if err != nil {
		return nil, err
	}

	return &NatsProvider{
		Conn: sc,
	}, nil
}
