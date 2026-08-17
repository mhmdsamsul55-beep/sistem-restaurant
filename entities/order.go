package entities

import "time"

type Order struct {
	ID               int64     `json:"ID"`
	UserID           int64     `json:"UserID"`
	NamaLengkap      string    `json:"NamaLengkap"`
	TotalHarga       float64   `json:"TotalHarga"`
	MetodePembayaran string    `json:"MetodePembayaran"`
	Status           string    `json:"Status"`
	CreatedAt        time.Time `json:"CreatedAt"`
}
