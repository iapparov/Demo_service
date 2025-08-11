package db

import (
	"context"
	"demoservice/internal/app"
	"demoservice/internal/config"
	"errors"
	"fmt"
	"log"
	"time"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface{
	Save(order *app.Order) error
	Load(uid string) (*app.Order, error)
	CacheLoad(conf *config.Config) ([]*app.Order, error)
}

type PostgresRepo struct {
	conn *pgxpool.Pool
}

func NewPostgresRepo(conn *pgxpool.Pool) *PostgresRepo{
	return &PostgresRepo{conn:conn}
}

func ConnectDB(conf *config.Config) *pgxpool.Pool{
	connStr := "postgres://"+conf.DbUser+":"+conf.DbPassword+"@"+conf.DbUrl+":"+conf.DbPort+"/"+conf.DbName
	context_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := pgxpool.New(context_, connStr)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	return conn
}

func (s *PostgresRepo) Save(order *app.Order) error{
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}

	query_order := `INSERT INTO orders (
		order_uid,
		track_number,
		entry,
		locale,
		internal_signature,
		customer_id,
		delivery_service,
		shardkey,
		sm_id,
		date_created,
		oof_shard
	  )
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`
	_, err = tx.Exec(ctx, query_order, order.OrderUid, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerId, order.DeliveryService, order.Shardkey, order.SmId, order.DateCreated, order.OofShard)
	if err != nil {
		log.Printf("Query_order trouble ")
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Printf("Rollback error: %v", rbErr)
		}
		return err
	}
	
	for _, elem := range order.Items{
		query_items := `INSERT INTO items (
			order_uid,
			chrt_id,
			track_number,
			price,
			rid,
			name,
			sale,
			size,
			total_price,
			nm_id,
			brand,
			status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`
		_, err = tx.Exec(ctx, query_items, order.OrderUid, elem.ChrtId, elem.TrackNumber, elem.Price, elem.Rid, elem.Name, elem.Sale, 
			elem.Size, elem.TotalPrice, elem.NmId, elem.Brand, elem.Status)
		if err != nil {
			log.Printf("Query_items trouble ")
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Printf("Rollback error: %v", rbErr)
			}
			return err
		}
	}

	
	p := &order.Payment
	p_dt := time.Unix(p.PaymentDt, 0)
	query_payment := `INSERT INTO payments (
		order_uid,
		transaction,
		request_id,
		currency,
		provider,
		amount,
		payment_dt,
		bank,
		delivery_cost,
		goods_total,
		custom_fee
	  )
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`
	_, err = tx.Exec(ctx, query_payment, order.OrderUid, p.Transaction, p.RequestId, p.Currency, p.Provider, p.Amount, p_dt, p.Bank, p.DeliveryCost, p.GoodsTotal, p.CustomFee)
	if err != nil {
		log.Printf("Query_payment trouble ")
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Printf("Rollback error: %v", rbErr)
		}
		return err
	}

	d := &order.Delivery
	query_deliveris := `INSERT INTO deliveries (
		order_uid,
		name,
		phone,
		zip,
		city,
		address,
		region,
		email
		)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`
	_, err = tx.Exec(ctx, query_deliveris, order.OrderUid, d.Name, d.Phone, d.Zip, d.City, d.Address, d.Region, d.Email)
	if err != nil {
		log.Printf("Query_deliveries trouble ")
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Printf("Rollback error: %v", rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
    	log.Printf("Commit error: %v", err)
	}
	return nil
}


func (s *PostgresRepo) Load(uid string) (*app.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var order app.Order
	order.OrderUid = uid

	if err := s.conn.QueryRow(ctx,
		`SELECT name, phone, zip, city, address, region, email
		 FROM deliveries WHERE order_uid = $1`, uid).
		Scan(
			&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
			&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
		); err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, fmt.Errorf("delivery not found for order_uid %s", uid)
    }
    return nil, err
}

	var paymentTime time.Time
	if err := s.conn.QueryRow(ctx,
		`SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		 FROM payments WHERE order_uid = $1`, uid).
		Scan(
			&order.Payment.Transaction, &order.Payment.RequestId, &order.Payment.Currency,
			&order.Payment.Provider, &order.Payment.Amount, &paymentTime,
			&order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee,
		); err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, fmt.Errorf("delivery not found for order_uid %s", uid)
    }
    return nil, err
}
	order.Payment.PaymentDt = paymentTime.Unix()

	if err := s.conn.QueryRow(ctx,
		`SELECT track_number, entry, locale, internal_signature, customer_id,
		        delivery_service, shardkey, sm_id, date_created, oof_shard
		 FROM orders WHERE order_uid = $1`, uid).
		Scan(
			&order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
			&order.CustomerId, &order.DeliveryService, &order.Shardkey, &order.SmId,
			&order.DateCreated, &order.OofShard,
		); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("delivery not found for order_uid %s", uid)
    }
    return nil, err
}
	rows, err := s.conn.Query(ctx,
		`SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		 FROM items WHERE order_uid = $1`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item app.Item
		if err := rows.Scan(
			&item.ChrtId, &item.TrackNumber, &item.Price, &item.Rid, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmId, &item.Brand, &item.Status,
		); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &order, nil
}


func (s *PostgresRepo) CacheLoad(conf *config.Config) ([]*app.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query_uids := `SELECT order_uid FROM orders ORDER BY date_created desc LIMIT $1`

	rows, err := s.conn.Query(ctx, query_uids, conf.CacheSize)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*app.Order{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	uids := make([]string, 0, conf.CacheSize)
	var uid string
	for rows.Next() {
		if err = rows.Scan(&uid); err != nil {
			return nil, err
		}
		uids = append(uids, uid)
	}

	Orders := make([]*app.Order, len(uids))

	for idx, elem := range uids {
		Orders[idx], err = s.Load(elem)

		if err != nil {
			return nil, err
		}
	}
	return Orders, nil
}