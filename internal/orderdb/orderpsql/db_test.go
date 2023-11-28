package orderpsql

import (
	"context"
	"encoding/json"
	"orderservice/internal/schema"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	order := schema.Order{OrderUID: "123142512", TrackNumber: "12314", Entry: "64363"}

	data, err := json.Marshal(&order)
	require.NoError(t, err)

	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mock.Close(context.Background())

	mock.ExpectExec("INSERT INTO orderDB (.+)").
		WithArgs(order.OrderUID, data).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	db := New(
		Config{QueryTimeout: time.Second},
		Dependencies{
			Log: logrus.StandardLogger(),
			PGX: mock,
		})

	err = db.AddOrder(context.Background(), order)
	require.NoError(t, err)
}
