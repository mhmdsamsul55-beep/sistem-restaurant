package entities

type OrderDetail struct {
	ID        int64
	OrderID   int64
	ProductID int64
	Jumlah    int
	Harga     float64
	Subtotal  float64
}