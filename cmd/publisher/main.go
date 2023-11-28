package main

import (
	"context"
	"orderservice/internal/orderevent"
	"orderservice/internal/orderevent/ordernats"
	"orderservice/internal/provider/natsprovider"
	"orderservice/internal/schema"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	log := logrus.New()

	countToPublish := 1
	if len(os.Args) > 1 {
		v, err := strconv.ParseInt(os.Args[1], 10, 32)
		if err != nil {
			log.Errorf("failed to parse cmd arg: %v", err)
			return
		}

		countToPublish = int(v)
	}

	if err := godotenv.Load(); err != nil {
		log.Errorf("Error loading .env file: %v", err)
	}

	np, err := natsprovider.New(natsprovider.Config{
		StanClusterID: os.Getenv("STAN_CLUSTER_ID"),
		ClientID:      "user2",
		URL:           os.Getenv("NATS_URL"),
	})
	if err != nil {
		log.Errorf("failed create nats provider: %v", err)
		return
	}
	defer np.Close()

	ordernats := ordernats.New(
		ordernats.Config{
			ChannelName: os.Getenv("STAN_CHANNEL_NAME"),
		},
		ordernats.Dependencies{
			Log:        log,
			NSProvider: np,
		})
	if err != nil {
		log.Errorf("failed to created ordernats: %v", err)
		return
	}

	if err := Publish(ctx, log, ordernats, countToPublish); err != nil {
		log.Errorf("failed to publish: %v", err)
	}
}

func Publish(ctx context.Context, log *logrus.Logger, p orderevent.OrderPublisher, count int) error {
	for i := 0; i < count; i++ {
		uid := uuid.NewString()
		order := schema.Order{
			OrderUID:    schema.OrderUID(uid),
			TrackNumber: "WBILMTESTTRACK",
			Entry:       "WBIL",
			Delivery: schema.Delivery{
				Name:   "Test Testov",
				Phone:  "+9720000000",
				Zip:    2639809,
				City:   "Kiryat Mozkin",
				Adress: "Ploshad Mira 15",
				Region: "Kraiot",
				Email:  "test@gmail.com",
			},
			Payment: schema.Payment{
				Transaction:   "b563feb7b2b84b6test",
				RequestID:     "",
				Currency:      "USD",
				Provider:      "wbpay",
				Amount:        1817,
				PaymentDT:     1637907727,
				Bank:          "alpha",
				DeliveryConst: 1500,
				GoodsTotal:    317,
			},
			Items: schema.Items{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				RID:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        0,
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
			Locale:          "en",
			InternalSign:    "",
			CustomerID:      "test",
			DeliveryService: "meest",
			Shardkey:        9,
			SmID:            99,
			DateCreated:     "2021-11-26T06:22:19Z",
			OofShard:        1,
		}

		if err := p.PublishOrder(ctx, order); err != nil {
			log.Errorf("failed to publish: %s(%d)", order.OrderUID, i)
			return err
		}

		log.Infof("order published: %s", order.OrderUID)
	}

	return nil
}
