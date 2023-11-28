package main

import (
	"context"
	"errors"
	"orderservice/internal/orderdb/ordercache"
	postgres "orderservice/internal/orderdb/orderpsql"
	"orderservice/internal/orderevent/ordernats"
	"orderservice/internal/provider/natsprovider"
	"orderservice/internal/provider/pgxprovider"
	"orderservice/internal/server"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log := logrus.New()
	if err := godotenv.Load(); err != nil {
		log.Errorf("error loading .env file: %v", err)
		return
	}

	pgxp, err := pgxprovider.New(pgxprovider.Config{
		URL: os.Getenv("POSTGRES_URL"),
	})
	if err != nil {
		log.Errorf("failed postges: %v", err)
		return
	}
	defer pgxp.Close(context.Background())

	db := postgres.New(
		postgres.Config{
			QueryTimeout: 1 * time.Second,
		},
		postgres.Dependencies{
			Log: log,
			PGX: pgxp,
		})

	cache := ordercache.New(
		ordercache.Config{},
		ordercache.Dependencies{
			Persistent: db,
		})

	np, err := natsprovider.New(natsprovider.Config{
		StanClusterID: os.Getenv("STAN_CLUSTER_ID"),
		ClientID:      "user1",
		URL:           os.Getenv("NATS_URL"),
	})
	if err != nil {
		log.Errorf("failed to create nats provider: %v", err)
		return
	}

	eventConsumer := ordernats.New(
		ordernats.Config{
			ChannelName: os.Getenv("STAN_CHANNEL_NAME"),
			QueueDepth:  1024,
		},
		ordernats.Dependencies{
			Log:        log,
			NSProvider: np,
			Store:      cache,
		})
	if err := eventConsumer.SubscribeOnOrder(); err != nil {
		log.Errorf("failed to subscribe on order: %v", err)
		return
	}
	defer eventConsumer.Unsubscribe()

	server := server.NewServer(
		server.Config{Address: os.Getenv("SERVER_ADDR")},
		server.Dependencies{
			Log: log,
			DB:  cache,
		})
	if err = cache.Restore(ctx); err != nil {
		log.Errorf("cache restore error: %v", err)
		return
	}

	if err = server.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Errorf("failed to run server: %v", err)
	}
}
