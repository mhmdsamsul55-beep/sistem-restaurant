package entities

import "time"

type User struct {
	ID          int64
	NamaLengkap string
	Email       string
	NomerHP     string
	Password    string
	KonfirmasiPassword string
	Role        string
	Provider    string
	Picture     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}