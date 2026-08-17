package controller

import (
	"Auth/entities"
	"encoding/json"
	"net/http"
	"strconv"
)

func GetAllOrdersAdmin(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}
	orders, err := OrderModel.GetAllOrders()
	if err != nil {
		http.Error(w, "Gagal mengambil pesanan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(orders)
}

func UpdateOrderStatusAdmin(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPut {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Data tidak valid", http.StatusBadRequest)
		return
	}
	if input.ID <= 0 {
		http.Error(w, "ID pesanan tidak valid", http.StatusBadRequest)
		return
	}
	switch input.Status {
	case "menunggu":
	case "diproses":
	case "siap":
	case "selesai":
	case "dibatalkan":
	default:
		http.Error(w, "Status tidak valid", http.StatusBadRequest)
		return
	}
	err = OrderModel.UpdateStatus(input.ID, input.Status)
	if err != nil {
		http.Error(w, "Gagal mengubah status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Status pesanan berhasil diubah",
	})
}
var _ = strconv.IntSize
var _ entities.Order
