package orderevent

import (
	"context"
	"orderservice/internal/schema"
)

type OrderPublisher interface {
	PublishOrder(context.Context, schema.Order) error
}

type OrderConsumer interface {
	SubscribeOnOrder(context.Context) error
	Unsubscribe()
}
