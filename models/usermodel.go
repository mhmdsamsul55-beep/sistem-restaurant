package models

import (
	"Auth/config"
	"Auth/entities"
	"database/sql"
	"fmt"
)

type UserModel struct {
	db *sql.DB
}

// ========================================
// KONEKSI DATABASE
// ========================================

func NewUserModel() *UserModel {

	conn, err := config.DBconn()

	if err != nil {
		panic(err)
	}

	return &UserModel{
		db: conn,
	}
}

// ========================================
// CREATE USER
// Dipakai untuk register biasa
// dan otomatis membuat user OAuth
// ========================================

func (u UserModel) Create(user *entities.User) error {

	fmt.Println("===== DATA USER OAUTH =====")
	fmt.Println("Nama       :", user.NamaLengkap)
	fmt.Println("Email      :", user.Email)
	fmt.Println("Nomor HP   :", user.NomerHP)
	fmt.Println("Password   :", user.Password)
	fmt.Println("Konfirmasi :", user.KonfirmasiPassword)
	fmt.Println("Role       :", user.Role)
	fmt.Println("Provider   :", user.Provider)
	fmt.Println("Picture    :", user.Picture)
	fmt.Println("===========================")

	query := `
		INSERT INTO userss (
			nama_lengkap,
			nomor_hp,
			email,
			password,
			konfirmasi_password,
			role,
			provider,
			picture
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := u.db.Exec(
		query,
		user.NamaLengkap,
		user.NomerHP,
		user.Email,
		user.Password,
		user.KonfirmasiPassword,
		user.Role,
		user.Provider,
		user.Picture,
	)

	return err
}

// ========================================
// WHERE
// ========================================

func (u UserModel) Where(
	user *entities.User,
	fieldName string,
	fieldValue interface{},
) error {

	allowedFields := map[string]bool{
		"id":       true,
		"email":    true,
		"nomor_hp": true,
	}

	if !allowedFields[fieldName] {
		return sql.ErrNoRows
	}

	query := `
		SELECT
			id,
			nama_lengkap,
			email,
			nomor_hp,
			password,
			konfirmasi_password,
			role,
			provider,
			picture,
			created_at,S
			updated_at
		FROM userss
		WHERE ` + fieldName + ` = ?
		LIMIT 1
	`

	err := u.db.QueryRow(
		query,
		fieldValue,
	).Scan(
		&user.ID,
		&user.NamaLengkap,
		&user.Email,
		&user.NomerHP,
		&user.Password,
		&user.KonfirmasiPassword,
		&user.Role,
		&user.Provider,
		&user.Picture,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	return err
}

// ========================================
// CEK DATA SUDAH ADA ATAU BELUM
// ========================================

func (u UserModel) Exist(
	fieldName string,
	fieldValue string,
) bool {

	allowedFields := map[string]bool{
		"id":       true,
		"email":    true,
		"nomor_hp": true,
	}

	if !allowedFields[fieldName] {
		return false
	}

	query := `
		SELECT COUNT(*)
		FROM userss
		WHERE ` + fieldName + ` = ?
	`

	var count int

	err := u.db.QueryRow(
		query,
		fieldValue,
	).Scan(&count)

	if err != nil {
		return false
	}

	return count > 0
}

// ========================================
// LOGIN DENGAN EMAIL / NOMOR HP
// ========================================

func (u UserModel) GetByID(
	id int64,
) (*entities.User, error) {

	var user entities.User

	query := `
		SELECT
			id,
			nama_lengkap,
			email,
			nomor_hp,
			password,
			konfirmasi_password,
			role,
			provider,
			picture
		FROM userss
		WHERE id = ?
	`

	err := u.db.QueryRow(
		query,
		id,
	).Scan(
		&user.ID,
		&user.NamaLengkap,
		&user.Email,
		&user.NomerHP,
		&user.Password,
		&user.KonfirmasiPassword,
		&user.Role,
		&user.Provider,
		&user.Picture,
	
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ========================================
// UPDATE PROFIL
// ========================================

func (u UserModel) Update(
	user *entities.User,
) error {

	query := `
		UPDATE userss
		SET
			nama_lengkap = ?,
			email = ?,
			nomor_hp = ?,
			picture = ?
		WHERE id = ?
	`

	_, err := u.db.Exec(
		query,
		user.NamaLengkap,
		user.Email,
		user.NomerHP,
		user.Picture,
		user.ID,
	)

	return err
}

// ========================================
// UPDATE ROLE
// ========================================

func (u UserModel) UpdateRole(
	id int64,
	role string,
) error {

	if role != "user" && role != "admin" {
		return sql.ErrNoRows
	}

	query := `
		UPDATE userss
		SET role = ?
		WHERE id = ?
	`

	_, err := u.db.Exec(
		query,
		role,
		id,
	)

	return err
}

// ========================================
// DELETE USER
// ========================================

func (u UserModel) Delete(
	id int64,
) error {

	query := `
		DELETE FROM userss
		WHERE id = ?
	`

	_, err := u.db.Exec(
		query,
		id,
	)

	return err
}

// ========================================
// GET ALL USER
// ========================================

func (u UserModel) GetAll() ([]entities.User, error) {

	query := `
		SELECT
			id,
			nama_lengkap,
			email,
			nomor_hp,
			password,
			konfirmasi_password,
			role,
			provider,
			picture,
			created_at,
			updated_at
		FROM userss
		ORDER BY id DESC
	`

	rows, err := u.db.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var userss []entities.User

	for rows.Next() {

		var user entities.User

		err := rows.Scan(
			&user.ID,
			&user.NamaLengkap,
			&user.Email,
			&user.NomerHP,
			&user.Password,
			&user.KonfirmasiPassword,
			&user.Role,
			&user.Provider,
			&user.Picture,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		userss = append(userss, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return userss, nil
}

func (u UserModel) GetByLogin(login string) (*entities.User, error) {

	var user entities.User

	query := `
		SELECT
			id,
			nama_lengkap,
			email,
			nomor_hp,
			password,
			konfirmasi_password,
			role,
			provider,
			picture
		FROM userss
		WHERE email = ?
		   OR nomor_hp = ?
		LIMIT 1
	`

	err := u.db.QueryRow(
		query,
		login,
		login,
	).Scan(
		&user.ID,
		&user.NamaLengkap,
		&user.Email,
		&user.NomerHP,
		&user.Password,
		&user.KonfirmasiPassword,
		&user.Role,
		&user.Provider,
		&user.Picture,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
func (u *UserModel) FindByEmail(email string) (entities.User, error) {
	var user entities.User

	query := `
		SELECT
			id,
			nama_lengkap,
			email,
			nomor_hp,
			password,
			konfirmasi_password,
			role,
			provider,
			picture
		FROM userss
		WHERE email = ?
		LIMIT 1
	`

	err := u.db.QueryRow(
		query,
		email,
	).Scan(
		&user.ID,
		&user.NamaLengkap,
		&user.Email,
		&user.NomerHP,
		&user.Password,
		&user.KonfirmasiPassword,
		&user.Role,
		&user.Provider,
		&user.Picture,
	)

	return user, err
}
