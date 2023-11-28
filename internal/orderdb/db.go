package orderdb

import (
	"context"
	"errors"
	"orderservice/internal/schema"
)

var ErrNotFound = errors.New("not found")

//go:generate mockgen -package orderdb -destination db_mock.go . OrderDB
type OrderDB interface {
	SeqNumber(ctx context.Context) (schema.SeqNumber, error)
	AddOrder(ctx context.Context, order schema.Order, seq schema.SeqNumber) error
	GetOrder(ctx context.Context, orderUI schema.OrderUID) (schema.Order, error)
	ListOrders(ctx context.Context) ([]schema.Order, error)
}

type RestorableOrderDB interface {
	Restore(ctx context.Context) error
}
