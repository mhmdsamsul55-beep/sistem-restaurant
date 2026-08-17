package models

import (
	"Auth/config"
	"Auth/entities"
	"database/sql"
	"fmt"
)

type ProductsModel struct {
	db *sql.DB
}

// koneksi ke database
func NewProductsModel() *ProductsModel {

	conn, err := config.DBconn()

	if err != nil {
		panic(err)
	}

	return &ProductsModel{
		db: conn,
	}

}

func (p *ProductsModel) Create(products entities.Products) error {
	query := `INSERT INTO products (nama_produk, deskripsi, 
   harga, stok, kategori,
    gambar, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := p.db.Exec(
		query,
		products.NamaProduk,
		products.Deskripsi,
		products.Harga,
		products.Stok,
		products.Kategori,
		products.Gambar,
		products.CreatedAt,
		products.UpdatedAt,
	)
	return err
}

func (p *ProductsModel) GetAll() ([]entities.Products, error) {
	query := `SELECT id, nama_produk,deskripsi, harga, stok, kategori, gambar, created_at, updated_at FROM products`
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err

	}
	defer rows.Close()

	var ProductsList []entities.Products
	for rows.Next() {

		var produk entities.Products

		if err := rows.Scan(&produk.ID,
			&produk.NamaProduk,
			&produk.Deskripsi,
			&produk.Harga,
			&produk.Stok,
			&produk.Kategori,
			&produk.Gambar,
			&produk.CreatedAt,
			&produk.UpdatedAt,
		); err != nil {
			return nil, err
		}

		ProductsList = append(ProductsList, produk)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ProductsList, nil
}

func (p *ProductsModel) Search(keyword string, produk *[]entities.Products) error {
	query := `SELECT
        id,
        nama_produk,
        harga,
        stok,
        deskripsi,
        gambar,
        created_at
    FROM products`
	args := []interface{}{}

	if keyword != "" {
		query += ` WHERE nama_produk LIKE ? OR deskripsi LIKE ?`
		pattern := "%" + keyword + "%"
		args = append(args, pattern, pattern)
	}

	rows, err := p.db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var pitem entities.Products
		if err := rows.Scan(&pitem.ID, &pitem.NamaProduk, &pitem.Harga, &pitem.Stok,
			&pitem.Deskripsi, &pitem.Gambar, &pitem.CreatedAt); err != nil {
			return err
		}
		*produk = append(*produk, pitem)
	}

	return rows.Err()

}

func (p *ProductsModel) Delete(id int64) error {

	query := `
		DELETE FROM products
		WHERE id = ?
	`

	result, err := p.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("produk dengan ID %d tidak ditemukan", id)
	}

	return nil
}
func (m *ProductsModel) KurangiStok(id int, jumlah int) error {

	query := `
		UPDATE products
		SET stok = stok - ?
		WHERE id = ?
		AND stok >= ?
	`

	result, err := m.db.Exec(
		query,
		jumlah,
		id,
		jumlah,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("stok tidak mencukupi atau produk tidak ditemukan")
	}

	return nil
}
