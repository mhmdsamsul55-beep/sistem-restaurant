package entities

import "time"

type Payment struct {
	ID          int64
	OrderID     int64
	Metode      string
	Status      string
	JumlahBayar float64
	CreatedAt   time.Time
}