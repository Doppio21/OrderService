package pgxprovider

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	defaultConnectTimeout = 2 * time.Second
)

type PGXInterface interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Close(ctx context.Context) error
}

type Config struct {
	URL            string
	ConnectTimeout time.Duration
}

type PGXProvider struct {
	*pgx.Conn
}

func New(cfg Config) (*PGXProvider, error) {
	connectTimeout := defaultConnectTimeout
	if cfg.ConnectTimeout != 0 {
		connectTimeout = cfg.ConnectTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	conn, err := pgx.Connect(ctx, cfg.URL)
	if err != nil {
		return nil, err
	}

	return &PGXProvider{
		Conn: conn,
	}, nil
}

func (pgx *PGXProvider) Close(ctx context.Context) error {
	return pgx.Conn.Close(context.Background())
}
