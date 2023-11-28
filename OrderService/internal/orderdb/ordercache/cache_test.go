package ordercache

import (
	"context"
	"orderservice/internal/orderdb"
	"orderservice/internal/schema"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var testOrder = []schema.Order{
	{OrderUID: "1234", Entry: "64363"},
	{OrderUID: "2234", Entry: "64363"},
}

func TestAddGet(t *testing.T) {
	testOrder := schema.Order{TrackNumber: "12314", Entry: "64363"}

	type test struct {
		name    string
		setUID  schema.OrderUID
		getUID  schema.OrderUID
		wantErr error
		setup   func(*orderdb.MockOrderDB)
	}

	cases := []test{
		{
			name:    "not_found",
			setUID:  "key1",
			getUID:  "key2",
			wantErr: orderdb.ErrNotFound,
			setup: func(db *orderdb.MockOrderDB) {
				order := testOrder
				order.OrderUID = "key1"
				db.EXPECT().AddOrder(gomock.Any(), order)
			},
		},
		{
			name:   "success",
			setUID: "key1",
			getUID: "key1",
			setup: func(db *orderdb.MockOrderDB) {
				order := testOrder
				order.OrderUID = "key1"
				db.EXPECT().AddOrder(gomock.Any(), order)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := orderdb.NewMockOrderDB(ctrl)
			if c.setup != nil {
				c.setup(db)
			}

			cache := New(Config{}, Dependencies{
				Persistent: db,
			})

			order := testOrder
			order.OrderUID = c.setUID
			err := cache.AddOrder(context.Background(), order)
			require.NoError(t, err)

			gotOrder, err := cache.GetOrder(context.Background(), c.getUID)
			if c.wantErr != nil {
				require.ErrorIs(t, err, c.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, order, gotOrder)
		})
	}
}

func TestList(t *testing.T) {
	type test struct {
		name  string
		set   []schema.Order
		list  []schema.Order
		setup func(*orderdb.MockOrderDB)
	}

	cases := []test{
		{
			name: "success",
			set:  testOrder,
			list: testOrder,
			setup: func(db *orderdb.MockOrderDB) {
				for _, order := range testOrder {
					db.EXPECT().AddOrder(gomock.Any(), order)
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := orderdb.NewMockOrderDB(ctrl)
			if c.setup != nil {
				c.setup(db)
			}

			cache := New(Config{}, Dependencies{
				Persistent: db,
			})

			for _, s := range c.set {
				err := cache.AddOrder(context.Background(), s)
				require.NoError(t, err)
			}

			res, err := cache.ListOrders(context.Background())
			require.NoError(t, err)
			require.Len(t, res, len(testOrder))
			for _, order := range res {
				require.Contains(t, testOrder, order)
			}
		})
	}
}

func TestRestore(t *testing.T) {
	ctrl := gomock.NewController(t)

	db := orderdb.NewMockOrderDB(ctrl)
	db.EXPECT().ListOrders(gomock.Any()).Return(testOrder, nil)

	cache := New(
		Config{},
		Dependencies{
			Persistent: db,
		})

	err := cache.Restore(context.Background())
	require.NoError(t, err)

	res, err := cache.ListOrders(context.Background())
	require.NoError(t, err)
	require.Len(t, res, len(testOrder))
	for _, order := range res {
		require.Contains(t, testOrder, order)
	}
}
