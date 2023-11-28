package ordernats

import (
	"context"
	"encoding/json"
	"errors"
	"orderservice/internal/orderdb"
	"orderservice/internal/provider/natsprovider"
	"orderservice/internal/schema"

	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
)

type Config struct {
	QueueDepth  int
	ChannelName string
}

type Dependencies struct {
	Log        *logrus.Logger
	NSProvider *natsprovider.NatsProvider
	Store      orderdb.OrderDB
}

type NatsOrderStore struct {
	cfg  Config
	deps Dependencies

	sub *stan.Subscription
	log *logrus.Entry
}

func New(cfg Config, deps Dependencies) *NatsOrderStore {
	return &NatsOrderStore{
		cfg:  cfg,
		deps: deps,
		log:  deps.Log.WithField("component", "ordernats"),
	}
}

func (n *NatsOrderStore) PublishOrder(ctx context.Context, order schema.Order) error {
	data, err := json.Marshal(&order)
	if err != nil {
		return err
	}

	err = n.deps.NSProvider.Publish(n.cfg.ChannelName, data)
	if err != nil {
		n.log.Errorf("failed to publish order: %v", err)
		return err
	}

	n.log.Info("order published sucessfully")
	return nil
}

func (n *NatsOrderStore) SubscribeOnOrder(ctx context.Context) error {
	if n.sub != nil {
		return errors.New("already subscribed")
	}

	seq, err := n.deps.Store.SeqNumber(ctx)
	if err != nil {
		return err
	}

	sub, err := n.deps.NSProvider.Subscribe(n.cfg.ChannelName, func(msg *stan.Msg) {
		order := schema.Order{}
		if err := json.Unmarshal(msg.Data, &order); err != nil {
			n.log.Errorf("invalid order scheme: %v", err)
			return
		}

		if err := n.deps.Store.AddOrder(context.Background(), order,
			schema.SeqNumber(msg.Sequence)); err != nil {
			return
		}

		// Операция вставки в БД идемпотента
		// Поэтому не имеет смысла обрабатывать ошибочное подтверждение обработки сообщения
		if err := msg.Ack(); err != nil {
			n.log.Errorf("failed to ack message: %v", err)
		}
	},
		stan.SetManualAckMode(),
		stan.MaxInflight(n.cfg.QueueDepth),
		stan.StartAtSequence(uint64(seq+1)))

	if err != nil {
		return err
	}

	n.sub = &sub
	return nil
}

func (n *NatsOrderStore) Unsubscribe() {
	if err := (*n.sub).Unsubscribe(); err != nil {
		n.log.Errorf("failed to unsubscribe: %v", err)
	}
}
