package models

import (
	"Auth/config"
	"Auth/entities"
	"database/sql"
	"time"
)

type OrderModel struct {
	db *sql.DB
}

func NewOrderModel() *OrderModel {

	conn, err := config.DBconn()

	if err != nil {
		panic(err)
	}

	return &OrderModel{
		db: conn,
	}
}
func (m *OrderModel) Create(order *entities.Order) error {

	result, err := m.db.Exec(`
		INSERT INTO orders
		(
			user_id,
			total_harga,
			metode_pembayaran,
			status
		)
		VALUES (?, ?, ?, ?)
	`,
		order.UserID,
		order.TotalHarga,
		order.MetodePembayaran,
		order.Status,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return err
	}

	order.ID = id

	return nil
}
func (m *OrderModel) GetByUserID(userID int64) ([]entities.Order, error) {

	rows, err := m.db.Query(`
		SELECT
			id,
			user_id,
			total_harga,
			metode_pembayaran,
			status,
			created_at
		FROM orders
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var orders []entities.Order

	for rows.Next() {

		var order entities.Order
		var createdAt []byte

		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.TotalHarga,
			&order.MetodePembayaran,
			&order.Status,
			&createdAt,
		)

		if err != nil {
			return nil, err
		}

		if len(createdAt) > 0 {

			order.CreatedAt, err = time.Parse(
				"2006-01-02 15:04:05",
				string(createdAt),
			)

			if err != nil {
				return nil, err
			}
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
func (m *OrderModel) GetAllOrders() ([]entities.Order, error) {

	rows, err := m.db.Query(`
		SELECT
			o.id,
			o.user_id,
			u.nama_lengkap,
			o.total_harga,
			o.metode_pembayaran,
			o.status,
			o.created_at
		FROM orders o
		LEFT JOIN userss u
			ON o.user_id = u.id
		ORDER BY o.created_at DESC
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var orders []entities.Order

	for rows.Next() {

		var order entities.Order
		var namaLengkap string
		var createdAt []byte

		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&namaLengkap,
			&order.TotalHarga,
			&order.MetodePembayaran,
			&order.Status,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}
		// Simpan nama pelanggan
		order.NamaLengkap = namaLengkap
		// Convert created_at
		if len(createdAt) > 0 {
			order.CreatedAt, err = time.Parse(
				"2006-01-02 15:04:05",
				string(createdAt),
			)

			if err != nil {
				return nil, err
			}
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}
func (m *OrderModel) UpdateStatus(id int64, status string) error {

	_, err := m.db.Exec(`
		UPDATE orders
		SET status = ?
		WHERE id = ?
	`,
		status,
		id,
	)

	return err
}
