package controller

import (
	"Auth/config"
	"Auth/entities"
	productsmodel "Auth/models"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var ProductsModel = productsmodel.NewProductsModel()

func Produk(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// Ambil semua produk dari database
	produk, err := ProductsModel.GetAll()
	if err != nil {
		http.Error(
			w,
			"Gagal mengambil produk: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	temp, err := template.ParseFiles("views/admin.html")
	if err != nil {
		http.Error(
			w,
			"Gagal membuka admin.html: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	// Data yang dikirim ke template
	data := map[string]interface{}{
		"products": produk,
	}

	err = temp.Execute(w, data)
	if err != nil {
		http.Error(
			w,
			"Gagal menampilkan admin: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}
}
func Tambah(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodGet:

		temp, err := template.ParseFiles("views/admin.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = temp.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		err := r.ParseMultipartForm(10 << 20) // maksimal 10 MB
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		file, handler, err := r.FormFile("gambar")
		if err != nil {
			http.Error(w, "Gambar tidak ditemukan: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		err = os.MkdirAll("assets/images", os.ModePerm)
		if err != nil {
			http.Error(w, "Gagal membuat folder gambar: "+err.Error(), http.StatusInternalServerError)
			return
		}
		dst, err := os.Create(
			filepath.Join("assets/images", handler.Filename),
		)
		if err != nil {
			http.Error(w, "Gagal menyimpan gambar: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, "Gagal menyalin gambar: "+err.Error(), http.StatusInternalServerError)
			return
		}
		namaProduk := r.Form.Get("nama_produk")
		deskripsi := r.Form.Get("deskripsi")
		kategori := r.Form.Get("kategori")

		harga, err := strconv.ParseFloat(r.Form.Get("harga"), 64)
		if err != nil {
			http.Error(w, "Harga tidak valid", http.StatusBadRequest)
			return
		}

		stok, err := strconv.Atoi(r.Form.Get("stok"))
		if err != nil {
			http.Error(w, "Stok tidak valid", http.StatusBadRequest)
			return
		}
		if kategori != "makanan" &&
			kategori != "minuman" &&
			kategori != "dessert" {

			http.Error(w, "Kategori tidak valid", http.StatusBadRequest)
			return
		}
		produk := entities.Products{
			NamaProduk: namaProduk,
			Deskripsi:  deskripsi,
			Harga:      harga,
			Stok:       stok,
			Kategori:   kategori,
			Gambar:     handler.Filename,
			CreatedAt:  time.Now().Format("2006-01-02 15:04:05"),
			UpdatedAt:  time.Now().Format("2006-01-02 15:04:05"),
		}

		idStr := r.FormValue("id")

		productID, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "ID produk tidak valid", http.StatusBadRequest)
			return
		}
		jumlahStr := r.FormValue("jumlah")

		jumlah, err := strconv.Atoi(jumlahStr)
		if err != nil || jumlah <= 0 {
			http.Error(w, "Jumlah tidak valid", http.StatusBadRequest)
			return
		}
		err = ProductsModel.Create(produk)
		if err != nil {
			http.Error(w, "Gagal menyimpan produk: "+err.Error(), http.StatusInternalServerError)
			return
		}
		err = ProductsModel.KurangiStok(productID, jumlah)
		if err != nil {
			// tampilkan error
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	default:
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
	}
}
func Search(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	session, _ := config.Store.Get(
		r,
		config.SESSION_ID,
	)

	if session.Values["LoggedIn"] != true {
		http.Redirect(
			w,
			r,
			"/login",
			http.StatusSeeOther,
		)
		return
	}

	keyword := r.URL.Query().Get("q")

	var produk []entities.Products

	err := ProductsModel.Search(
		keyword,
		&produk,
	)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	temp, err := template.ParseFiles(
		"views/admin.html",
	)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	data := map[string]interface{}{
		"nama_lengkap": session.Values["nama_lengkap"],
		"products":     produk,
		"keyword":      keyword,
	}

	err = temp.Execute(w, data)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}
}
func HapusProduk(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Method tidak diizinkan",
			http.StatusMethodNotAllowed,
		)
		return
	}

	// Ambil ID dari form
	idString := r.FormValue("id")

	if idString == "" {
		http.Error(
			w,
			"ID produk tidak ditemukan",
			http.StatusBadRequest,
		)
		return
	}

	// Ubah string menjadi int64
	id, err := strconv.ParseInt(idString, 10, 64)

	if err != nil {
		http.Error(
			w,
			"ID produk tidak valid",
			http.StatusBadRequest,
		)
		return
	}

	// Hapus produk
	err = ProductsModel.Delete(id)

	if err != nil {
		http.Error(
			w,
			"Gagal menghapus produk: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	// Kembali ke halaman admin
	http.Redirect(
		w,
		r,
		"/admin",
		http.StatusSeeOther,
	)
}
