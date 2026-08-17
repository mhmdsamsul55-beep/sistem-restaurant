package controller

import (
	"Auth/config"
	"Auth/entities"
	"Auth/models"
	"encoding/json"
	"net/http"
	"strconv"
)

var OrderModel = models.NewOrderModel()
func CreateOrder(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	session, err := config.Store.Get(r, config.SESSION_ID)
	if err != nil {
		http.Error(w, "Gagal mengambil session", http.StatusInternalServerError)
		return
	}

	if session.Values["LoggedIn"] != true {
		http.Error(w, "Silakan login terlebih dahulu", http.StatusUnauthorized)
		return
	}
	idValue := session.Values["id"]

	var userID int64

	switch id := idValue.(type) {

	case int64:
		userID = id

	case int:
		userID = int64(id)

	case string:
		userID, err = strconv.ParseInt(id, 10, 64)
		if err != nil {
			http.Error(w, "ID user tidak valid", http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, "ID user tidak ditemukan di session", http.StatusUnauthorized)
		return
	}

	var input struct {
		TotalHarga float64 `json:"total_harga"`

		MetodePembayaran string `json:"metode_pembayaran"`

		Items []struct {
			ID     string  `json:"id"`
			Nama   string  `json:"nama"`
			Harga  float64 `json:"harga"`
			Jumlah int     `json:"jumlah"`
		} `json:"items"`
	}

	err = json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		http.Error(
			w,
			"Data pesanan tidak valid: "+err.Error(),
			http.StatusBadRequest,
		)
		return
	}


	if input.TotalHarga <= 0 {
		http.Error(w, "Total harga tidak valid", http.StatusBadRequest)
		return
	}

	if input.MetodePembayaran == "" {
		http.Error(w, "Metode pembayaran belum dipilih", http.StatusBadRequest)
		return
	}

	if len(input.Items) == 0 {
		http.Error(w, "Keranjang kosong", http.StatusBadRequest)
		return
	}

	order := entities.Order{
		UserID:           userID,
		TotalHarga:       input.TotalHarga,
		MetodePembayaran: input.MetodePembayaran,
		Status:           "menunggu",
	}

	err = OrderModel.Create(&order)

	if err != nil {
		http.Error(w, "Gagal menyimpan pesanan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	produkModel := models.NewProductsModel()
	for _, item := range input.Items {
		productID, err := strconv.Atoi(item.ID)
		if err != nil {
			http.Error(w, "ID produk tidak valid: "+item.ID, http.StatusBadRequest)
			return
		}
		if item.Jumlah <= 0 {
			http.Error(w, "Jumlah produk tidak valid", http.StatusBadRequest)
			return
		}
		err = produkModel.KurangiStok(
			productID,
			item.Jumlah,
		)
		if err != nil {
			http.Error(w, "Gagal mengurangi stok produk "+item.Nama+": "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	w.Header().Set(
		"Content-Type",
		"application/json",
	)
	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"success":  true,
			"message":  "Pesanan berhasil dibuat",
			"order_id": order.ID,
		},
	)
}
func GetUserOrders(w http.ResponseWriter, r *http.Request) {

	session, err := config.Store.Get(r, config.SESSION_ID)
	if err != nil {
		http.Error(w, "Gagal mengambil session", http.StatusInternalServerError)
		return
	}
	if session.Values["LoggedIn"] != true {
		http.Error(w, "Silakan login terlebih dahulu", http.StatusUnauthorized)
		return
	}
	idValue := session.Values["id"]
	var userID int64
	switch id := idValue.(type) {
	case int64:
		userID = id
	case int:
		userID = int64(id)
	case string:
		userID, err = strconv.ParseInt(id, 10, 64)
		if err != nil {
			http.Error(w, "ID user tidak valid", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "ID user tidak ditemukan", http.StatusUnauthorized)
		return
	}
	orders, err := OrderModel.GetByUserID(userID)

	if err != nil {
		http.Error(w, "Gagal mengambil pesanan: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}
