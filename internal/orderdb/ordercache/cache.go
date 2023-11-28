package ordercache

import (
	"context"
	"orderservice/internal/orderdb"
	"orderservice/internal/schema"
	"sync"
	"sync/atomic"
)

type Config struct {
}

type Dependencies struct {
	Persistent orderdb.OrderDB
}

type CacheDB struct {
	cfg  Config
	deps Dependencies

	cached      sync.Map
	cachedCount atomic.Int32
}

func New(cfg Config, deps Dependencies) *CacheDB {
	return &CacheDB{
		cfg:  cfg,
		deps: deps,
	}
}

func (c *CacheDB) AddOrder(ctx context.Context, order schema.Order) error {
	if err := c.deps.Persistent.AddOrder(ctx, order); err != nil {
		return err
	}

	c.cached.Store(order.OrderUID, order)
	c.cachedCount.Add(1)
	return nil
}

func (c *CacheDB) GetOrder(_ context.Context, orderUID schema.OrderUID) (schema.Order, error) {
	v, ok := c.cached.Load(orderUID)
	if !ok {
		return schema.Order{}, orderdb.ErrNotFound
	}

	return v.(schema.Order), nil
}

func (c *CacheDB) ListOrders(_ context.Context) ([]schema.Order, error) {
	ret := make([]schema.Order, 0, c.cachedCount.Load())
	c.cached.Range(func(_, value any) bool {
		ret = append(ret, value.(schema.Order))
		return true
	})

	return ret, nil
}

func (c *CacheDB) Restore(ctx context.Context) error {
	res, err := c.deps.Persistent.ListOrders(ctx)
	if err != nil {
		return err
	}

	for _, order := range res {
		c.cached.Store(order.OrderUID, order)
	}

	return nil
}
