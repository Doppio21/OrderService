package orderpsql

import (
	"context"
	"encoding/json"
	"errors"
	"orderservice/internal/orderdb"
	"orderservice/internal/provider/pgxprovider"
	"orderservice/internal/schema"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type Config struct {
	QueryTimeout time.Duration
}

type Dependencies struct {
	Log *logrus.Logger
	PGX pgxprovider.PGXInterface
}

type Postgres struct {
	cfg  Config
	deps Dependencies

	log *logrus.Entry
}

func New(cfg Config, deps Dependencies) orderdb.OrderDB {
	return &Postgres{
		cfg:  cfg,
		deps: deps,
		log:  deps.Log.WithField("component", "orderdb"),
	}
}

func (p *Postgres) AddOrder(ctx context.Context, order schema.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, p.cfg.QueryTimeout)
	defer cancel()

	_, err = p.deps.PGX.Exec(ctx, `INSERT INTO orderDB (order_uid, data)
		VALUES ($1, $2)`, order.OrderUID, data)
	if err != nil {
		p.log.Errorf("failed to insert: %v", err)
		return err
	}

	p.log.Infof("order added: %s", order.OrderUID)
	return nil
}

func (p *Postgres) GetOrder(ctx context.Context, orderUI schema.OrderUID) (schema.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, p.cfg.QueryTimeout)
	defer cancel()

	res, err := p.deps.PGX.Query(ctx, `SELECT data FROM orderDB
		WHERE order_uid = &1)`, orderUI)
	if errors.Is(err, pgx.ErrNoRows) {
		return schema.Order{}, orderdb.ErrNotFound
	} else if err != nil {
		p.log.Errorf("failed to select: %v", err)
		return schema.Order{}, err
	}

	var data []byte
	if err = res.Scan(&data); err != nil {
		return schema.Order{}, err
	}

	var order schema.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return schema.Order{}, err
	}

	return order, nil
}

func (p *Postgres) ListOrders(ctx context.Context) ([]schema.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, p.cfg.QueryTimeout)
	defer cancel()

	res, err := p.deps.PGX.Query(ctx, "SELECT data FROM orderDB")
	if err != nil {
		p.log.Errorf("failed to list: %v", err)
		return nil, err
	}
	defer res.Close()

	ret := make([]schema.Order, 0)
	for res.Next() {
		var (
			data  []byte
			order schema.Order
		)

		if err = res.Scan(&data); err != nil {
			p.log.Errorf("Scan failed: %v", err)
			return nil, err
		}

		if err := json.Unmarshal(data, &order); err != nil {
			return nil, err
		}

		ret = append(ret, order)
	}

	return ret, nil
}
