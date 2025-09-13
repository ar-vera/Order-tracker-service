package repository

import (
	"Order-tracker-service/internal/domain"
	"github.com/jmoiron/sqlx"
)

type OrderRepository interface {
	Create(order *domain.Order) error
	GetById(orderId string) (*domain.Order, error)
	GetAll() ([]*domain.Order, error)
}

type OrderRepos struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *OrderRepos {
	return &OrderRepos{db: db}
}

func (r *OrderRepos) Create(order *domain.Order) (err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Вставляем заказ и получаем его id
	var orderID int
	err = tx.QueryRow(`
		INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id`,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.ShardKey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	).Scan(&orderID)
	if err != nil {
		return err
	}

	// Вставляем delivery
	_, err = tx.Exec(`
		INSERT INTO delivery (order_id, name, phone, zip, city, address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		orderID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	)
	if err != nil {
		return err
	}

	// Вставляем payment
	_, err = tx.Exec(`
		INSERT INTO payment (order_id, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		orderID,
		order.Payment.Transaction,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDT,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)
	if err != nil {
		return err
	}

	// Вставляем items
	for _, item := range order.Items {
		_, err = tx.Exec(`
			INSERT INTO items (order_id, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
			orderID,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.RID,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *OrderRepos) GetById(orderUID string) (*domain.Order, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	var order domain.Order
	var orderID int

	// Получаем заказ и его id
	row := tx.QueryRow(`
		SELECT id, order_uid, track_number, entry, locale, internal_signature, customer_id, 
		       delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders WHERE order_uid = $1`, orderUID)

	err = row.Scan(
		&orderID,
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.ShardKey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	)
	if err != nil {
		return nil, err
	}

	// Получаем delivery
	row = tx.QueryRow(`
		SELECT name, phone, zip, city, address, region, email 
		FROM delivery WHERE order_id = $1`, orderID)

	err = row.Scan(
		&order.Delivery.Name,
		&order.Delivery.Phone,
		&order.Delivery.Zip,
		&order.Delivery.City,
		&order.Delivery.Address,
		&order.Delivery.Region,
		&order.Delivery.Email,
	)
	if err != nil {
		return nil, err
	}

	// Получаем payment
	row = tx.QueryRow(`
		SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM payment WHERE order_id = $1`, orderID)

	err = row.Scan(
		&order.Payment.Transaction,
		&order.Payment.RequestID,
		&order.Payment.Currency,
		&order.Payment.Provider,
		&order.Payment.Amount,
		&order.Payment.PaymentDT,
		&order.Payment.Bank,
		&order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal,
		&order.Payment.CustomFee,
	)
	if err != nil {
		return nil, err
	}

	// Получаем items
	rows, err := tx.Query(`
		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.Item
		err = rows.Scan(
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.RID,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		)
		if err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}

func (r *OrderRepos) GetAll() ([]*domain.Order, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Получаем список заказов
	rows, err := tx.Query(`
		SELECT id, order_uid, track_number, entry, locale, internal_signature, customer_id,
		       delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders`)
	if err != nil {
		return nil, err
	}

	type rawOrder struct {
		ID    int
		Order domain.Order
	}

	var rawOrders []rawOrder

	for rows.Next() {
		var ro rawOrder
		if err := rows.Scan(
			&ro.ID,
			&ro.Order.OrderUID,
			&ro.Order.TrackNumber,
			&ro.Order.Entry,
			&ro.Order.Locale,
			&ro.Order.InternalSignature,
			&ro.Order.CustomerID,
			&ro.Order.DeliveryService,
			&ro.Order.ShardKey,
			&ro.Order.SmID,
			&ro.Order.DateCreated,
			&ro.Order.OofShard,
		); err != nil {
			rows.Close()
			return nil, err
		}
		rawOrders = append(rawOrders, ro)
	}
	rows.Close() // важно закрыть перед новыми Query!

	var orders []*domain.Order

	// Теперь подгружаем delivery, payment, items для каждого заказа
	for _, ro := range rawOrders {
		orderID := ro.ID
		order := ro.Order

		// delivery
		row := tx.QueryRow(`SELECT name, phone, zip, city, address, region, email 
		                    FROM delivery WHERE order_id=$1`, orderID)
		if err := row.Scan(
			&order.Delivery.Name,
			&order.Delivery.Phone,
			&order.Delivery.Zip,
			&order.Delivery.City,
			&order.Delivery.Address,
			&order.Delivery.Region,
			&order.Delivery.Email,
		); err != nil {
			return nil, err
		}

		// payment
		row = tx.QueryRow(`SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		                   FROM payment WHERE order_id=$1`, orderID)
		if err := row.Scan(
			&order.Payment.Transaction,
			&order.Payment.RequestID,
			&order.Payment.Currency,
			&order.Payment.Provider,
			&order.Payment.Amount,
			&order.Payment.PaymentDT,
			&order.Payment.Bank,
			&order.Payment.DeliveryCost,
			&order.Payment.GoodsTotal,
			&order.Payment.CustomFee,
		); err != nil {
			return nil, err
		}

		// items
		itemRows, err := tx.Query(`SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		                           FROM items WHERE order_id=$1`, orderID)
		if err != nil {
			return nil, err
		}
		for itemRows.Next() {
			var item domain.Item
			if err := itemRows.Scan(
				&item.ChrtID,
				&item.TrackNumber,
				&item.Price,
				&item.RID,
				&item.Name,
				&item.Sale,
				&item.Size,
				&item.TotalPrice,
				&item.NmID,
				&item.Brand,
				&item.Status,
			); err != nil {
				itemRows.Close()
				return nil, err
			}
			order.Items = append(order.Items, item)
		}
		itemRows.Close()

		orders = append(orders, &order)
	}

	return orders, nil
}
