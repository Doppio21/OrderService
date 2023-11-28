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
	PGX *pgxprovider.PGXProvider
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

func (p *Postgres) SeqNumber(ctx context.Context) (schema.SeqNumber, error) {
	ctx, cancel := context.WithTimeout(ctx, p.cfg.QueryTimeout)
	defer cancel()

	res, err := p.deps.PGX.Query(ctx, `SELECT seq FROM seqDB WHERE id = 1`)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	} else if err != nil {
		p.log.Errorf("failed to select seq number: %v", err)
		return 0, err
	}
	defer res.Close()

	var seq schema.SeqNumber
	for res.Next() {
		if err = res.Scan(&seq); err != nil {
			p.log.Errorf("scan failed: %v", err)
			return 0, err
		}

		break
	}

	return seq, nil
}

func (p *Postgres) AddOrder(ctx context.Context, order schema.Order, seq schema.SeqNumber) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, p.cfg.QueryTimeout)
	defer cancel()

	txn, err := p.deps.PGX.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		p.log.Errorf("failed to create transaction: %v", err)
		return err
	}

	_, err = txn.Exec(ctx, `INSERT INTO orderDB (order_uid, data)
		VALUES ($1, $2)`, order.OrderUID, data)
	if err != nil {
		p.log.Errorf("failed to insert: %v", err)
		return err
	}

	_, err = txn.Exec(ctx, `UPDATE seqDB SET seq = $1 WHERE seq < $1`, seq)
	if err != nil {
		p.log.Errorf("failed to save seq number: %v", err)
		return err
	}

	if err := txn.Commit(ctx); err != nil {
		p.log.Errorf("failed to commit order transaction: %v", err)
		return err
	}

	p.log.Infof("order added: %s", order.OrderUID)
	return nil
}

func (p *Postgres) GetOrder(ctx context.Context, orderUID schema.OrderUID) (schema.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, p.cfg.QueryTimeout)
	defer cancel()

	res, err := p.deps.PGX.Query(ctx, `SELECT data FROM orderDB
		WHERE order_uid = &1)`, orderUID)
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
