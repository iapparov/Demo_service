package db_test

import (
	"demoservice/internal/app"
	"demoservice/internal/config"
	"demoservice/internal/db"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func testConfig() *config.Config {
	return &config.Config{
		DbUser:     os.Getenv("TEST_DB_USER"),
		DbPassword: os.Getenv("TEST_DB_PASSWORD"),
		DbUrl:      os.Getenv("TEST_DB_HOST"),
		DbPort:     os.Getenv("TEST_DB_PORT"),
		DbName:     os.Getenv("TEST_DB_NAME"),
		CacheSize:  2,
	}
} //TEST_DB_USER=user TEST_DB_PASSWORD=user TEST_DB_HOST=localhost TEST_DB_PORT=5433 TEST_DB_NAME=mydb go test ./internal/db/tests

func TestPostgresRepo_SaveAndLoad(t *testing.T) {
	conf := testConfig()
	conn := db.ConnectDB(conf)
	repo := db.NewPostgresRepo(conn)

	order := &app.Order{
		OrderUid:          "test_uid_123",
		TrackNumber:       "TN123",
		Entry:             "entry",
		Locale:            "ru",
		InternalSignature: "sig",
		CustomerId:        "cust1",
		DeliveryService:   "ds",
		Shardkey:          "9",
		SmId:              1,
		DateCreated:       time.Now(),
		OofShard:          "1",
		Items: []app.Item{
			{ChrtId: 1, TrackNumber: "TN123", Price: 100, Rid: "rid1", Name: "item1", Sale: 0, Size: "M", TotalPrice: 100, NmId: 1, Brand: "brand", Status: 1},
		},
		Payment: app.Payment{
			Transaction:  "tx1",
			RequestId:    "req1",
			Currency:     "RUB",
			Provider:     "prov",
			Amount:       100,
			PaymentDt:    time.Now().Unix(),
			Bank:         "bank",
			DeliveryCost: 10,
			GoodsTotal:   90,
			CustomFee:    0,
		},
		Delivery: app.Delivery{
			Name:    "Ivan",
			Phone:   "123456",
			Zip:     "000000",
			City:    "Moscow",
			Address: "Lenina 1",
			Region:  "Moscow",
			Email:   "ivan@test.com",
		},
	}

	// Save order
	err := repo.Save(order)
	assert.NoError(t, err)

	// Load order
	loaded, err := repo.Load(order.OrderUid)
	assert.NoError(t, err)
	assert.Equal(t, order.OrderUid, loaded.OrderUid)
	assert.Equal(t, order.TrackNumber, loaded.TrackNumber)
	assert.Equal(t, order.Delivery.Name, loaded.Delivery.Name)
	assert.Equal(t, order.Payment.Transaction, loaded.Payment.Transaction)
	assert.Len(t, loaded.Items, 1)
	assert.Equal(t, order.Items[0].Name, loaded.Items[0].Name)
}

func TestPostgresRepo_CacheLoad(t *testing.T) {
	conf := testConfig()
	conn := db.ConnectDB(conf)
	repo := db.NewPostgresRepo(conn)

	orders, err := repo.CacheLoad(conf)
	assert.NoError(t, err)
	assert.LessOrEqual(t, len(orders), conf.CacheSize)
	for _, o := range orders {
		assert.NotEmpty(t, o.OrderUid)
	}
}
