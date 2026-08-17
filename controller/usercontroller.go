package controller

import (
	"Auth/config"
	"Auth/entities"
	usermodel "Auth/models"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var UserModel = usermodel.NewUserModel()
func getUserIDFromSession(session *sessions.Session) (int64, error) {

	idValue, ok := session.Values["id"]
	if !ok {
		return 0, fmt.Errorf("ID user tidak ditemukan di session")
	}

	switch v := idValue.(type) {

	case string:
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, err
		}
		return id, nil

	case int64:
		return v, nil

	case int:
		return int64(v), nil

	case int32:
		return int64(v), nil

	case float64:
		return int64(v), nil

	default:
		return 0, fmt.Errorf("tipe ID tidak dikenali: %T", idValue)
	}
}
func Index(w http.ResponseWriter, r *http.Request) {

	session, err := config.Store.Get(r, config.SESSION_ID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if session.Values["LoggedIn"] != true {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Kalau admin masuk ke halaman user
	role, _ := session.Values["role"].(string)

	if role == "admin" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	products, err := ProductsModel.GetAll()
	if err != nil {
		http.Error(
			w,
			"Gagal mengambil produk: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	temp, err := template.ParseFiles("views/index.html")
	if err != nil {
		http.Error(
			w,
			"Gagal membuka index.html: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	data := map[string]interface{}{
		"nama_lengkap": session.Values["nama_lengkap"],
		"email":        session.Values["email"],
		"nomor_hp":     session.Values["nomor_hp"],
		"role":         session.Values["role"],
		"picture":      session.Values["picture"],
		"provider":     session.Values["provider"],
		"products":     products,
	}

	err = temp.Execute(w, data)
	if err != nil {
		http.Error(
			w,
			"Gagal menampilkan index: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}
}

// =====================================================
// ADMIN
// =====================================================

func Admin(w http.ResponseWriter, r *http.Request) {

	session, err := config.Store.Get(r, config.SESSION_ID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if session.Values["LoggedIn"] != true {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	role, ok := session.Values["role"].(string)

	if !ok || role != "admin" {
		http.Error(
			w,
			"Akses ditolak. Kamu bukan admin.",
			http.StatusForbidden,
		)
		return
	}

	products, err := ProductsModel.GetAll()

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

	data := map[string]interface{}{
		"nama_lengkap": session.Values["nama_lengkap"],
		"email":        session.Values["email"],
		"nomor_hp":     session.Values["nomor_hp"],
		"role":         session.Values["role"],
		"picture":      session.Values["picture"],
		"provider":     session.Values["provider"],
		"products":     products,
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
func Login(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {

		temp, err := template.ParseFiles("views/login.html")
		if err != nil {
			http.Error(
				w,
				"Gagal membuka login.html: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		temp.Execute(w, nil)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Method tidak diizinkan",
			http.StatusMethodNotAllowed,
		)
		return
	}

	login := r.FormValue("login")
	password := r.FormValue("password")

	// Cari user berdasarkan email / nomor HP
	user, err := UserModel.GetByLogin(login)

	if err != nil {
		temp, _ := template.ParseFiles("views/login.html")

		temp.Execute(w, map[string]string{
			"logError": "Email atau nomor HP tidak ditemukan",
		})

		return
	}

	// Cek password
	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)

	if err != nil {
		temp, _ := template.ParseFiles("views/login.html")

		temp.Execute(w, map[string]string{
			"passError": "Password salah",
		})

		return
	}

	// Ambil session
	session, err := config.Store.Get(r, config.SESSION_ID)

	if err != nil {
		http.Error(
			w,
			"Gagal mengambil session: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	// Bersihkan session lama
	session.Values = make(map[interface{}]interface{})

	// Simpan data user
	session.Values["LoggedIn"] = true

	// ID disimpan sebagai STRING
	session.Values["id"] = strconv.FormatInt(user.ID, 10)

	session.Values["nama_lengkap"] = user.NamaLengkap
	session.Values["email"] = user.Email
	session.Values["nomor_hp"] = user.NomerHP
	session.Values["role"] = user.Role
	session.Values["provider"] = user.Provider
	session.Values["picture"] = user.Picture

	err = session.Save(r, w)

	if err != nil {
		http.Error(
			w,
			"Gagal menyimpan session: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	// Kalau admin
	if user.Role == "admin" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// User biasa
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func Logout(w http.ResponseWriter, r *http.Request) {

	session, err := config.Store.Get(r, config.SESSION_ID)

	if err != nil {
		http.Error(
			w,
			"Gagal mengambil session: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	session.Values = make(map[interface{}]interface{})

	session.Options.MaxAge = -1

	err = session.Save(r, w)

	if err != nil {
		http.Error(
			w,
			"Gagal menghapus session: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func Register(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		temp, err := template.ParseFiles("views/register.html")
		if err != nil {
			http.Error(
				w,
				err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		temp.Execute(w, nil)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Method tidak diizinkan",
			http.StatusMethodNotAllowed,
		)
		return
	}

	r.ParseForm()

	user := entities.User{

		NamaLengkap:        r.Form.Get("nama_lengkap"),
		Email:              r.Form.Get("email"),
		NomerHP:            r.Form.Get("nomor_hp"),
		Password:           r.Form.Get("password"),
		KonfirmasiPassword: r.Form.Get("konfirmasi_password"),

		Provider: "local",
		Role:     "user",
	}

	data := map[string]interface{}{}

	hasError := false

	// Validasi nama
	if user.NamaLengkap == "" {

		data["nameError"] = "Nama lengkap tidak boleh kosong"

		hasError = true
	}

	// Validasi email
	if user.Email == "" {

		data["emailError"] = "Email tidak boleh kosong"

		hasError = true
	}

	// Validasi nomor HP
	if user.NomerHP == "" {

		data["nomerError"] = "Nomor HP tidak boleh kosong"

		hasError = true
	}

	// Validasi password
	if user.Password == "" {

		data["passwordError"] = "Password tidak boleh kosong"

		hasError = true
	}

	// Validasi konfirmasi
	if user.KonfirmasiPassword == "" {

		data["konfirmasiError"] =
			"Konfirmasi password tidak boleh kosong"

		hasError = true
	}

	// Cek email
	if user.Email != "" &&
		UserModel.Exist("email", user.Email) {

		data["emailError"] = "Email sudah digunakan"

		hasError = true
	}

	// Cek nomor HP
	if user.NomerHP != "" &&
		UserModel.Exist("nomor_hp", user.NomerHP) {

		data["nomerError"] = "Nomor HP sudah digunakan"

		hasError = true
	}

	// Cek password
	if user.Password != "" &&
		user.KonfirmasiPassword != "" &&
		user.Password != user.KonfirmasiPassword {

		data["passwordError"] = "Password tidak sama"

		hasError = true
	}

	if hasError {

		data["nama_lengkap"] = user.NamaLengkap
		data["nomor_hp"] = user.NomerHP
		data["email"] = user.Email

		temp, _ := template.ParseFiles(
			"views/register.html",
		)

		temp.Execute(w, data)

		return
	}

	// Hash password
	hashPassword, err := bcrypt.GenerateFromPassword(
		[]byte(user.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		http.Error(
			w,
			"Gagal melakukan hash password",
			http.StatusInternalServerError,
		)
		return
	}

	user.Password = string(hashPassword)

	// Simpan user
	err = UserModel.Create(&user)

	if err != nil {
		http.Error(
			w,
			"Gagal membuat user: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	// Setelah register → login
	http.Redirect(
		w,
		r,
		"/login",
		http.StatusSeeOther,
	)
}
func UpdateFoto(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Method tidak diizinkan",
			http.StatusMethodNotAllowed,
		)
		return
	}

	session, err := config.Store.Get(
		r,
		config.SESSION_ID,
	)

	if err != nil {
		http.Redirect(
			w,
			r,
			"/login",
			http.StatusSeeOther,
		)
		return
	}

	if session.Values["LoggedIn"] != true {
		http.Redirect(
			w,
			r,
			"/login",
			http.StatusSeeOther,
		)
		return
	}

	// Ambil ID user dari session
	id, err := getUserIDFromSession(session)

	if err != nil {
		http.Error(
			w,
			"ID user tidak valid: "+err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	if id <= 0 {
		http.Error(
			w,
			"ID user tidak valid",
			http.StatusBadRequest,
		)
		return
	}

	// Pastikan user memang ada
	user, err := UserModel.GetByID(id)

	if err != nil {
		http.Error(
			w,
			"User tidak ditemukan: "+err.Error(),
			http.StatusNotFound,
		)
		return
	}

	// Ambil file foto
	file, header, err := r.FormFile("foto")

	if err != nil {
		http.Error(
			w,
			"Foto tidak ditemukan",
			http.StatusBadRequest,
		)
		return
	}

	defer file.Close()

	// Buat folder uploads
	err = os.MkdirAll(
		"uploads",
		os.ModePerm,
	)

	if err != nil {
		http.Error(
			w,
			"Gagal membuat folder uploads",
			http.StatusInternalServerError,
		)
		return
	}

	// Ekstensi file
	ext := filepath.Ext(header.Filename)

	// Nama file
	filename := fmt.Sprintf(
		"profile_%d_%d%s",
		id,
		time.Now().UnixNano(),
		ext,
	)

	filePath := filepath.Join(
		"uploads",
		filename,
	)

	// Buat file
	newFile, err := os.Create(filePath)

	if err != nil {
		http.Error(
			w,
			"Gagal membuat file foto",
			http.StatusInternalServerError,
		)
		return
	}

	defer newFile.Close()

	// Copy foto
	_, err = newFile.ReadFrom(file)

	if err != nil {
		http.Error(
			w,
			"Gagal menyimpan foto",
			http.StatusInternalServerError,
		)
		return
	}

	// Path database
	picture := "uploads/" + filename

	// Update picture
	user.Picture = picture

	err = UserModel.Update(user)

	if err != nil {
		http.Error(
			w,
			"Gagal mengupdate foto: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	// Update session
	session.Values["picture"] = picture

	err = session.Save(r, w)

	if err != nil {
		http.Error(
			w,
			"Gagal menyimpan session: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/profil",
		http.StatusSeeOther,
	)
}
func UpdateProfil(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Method tidak diizinkan",
			http.StatusMethodNotAllowed,
		)
		return
	}

	session, err := config.Store.Get(
		r,
		config.SESSION_ID,
	)

	if err != nil {
		http.Redirect(
			w,
			r, "/login", http.StatusSeeOther)
		return
	}

	if session.Values["LoggedIn"] != true {
		http.Redirect(
			w, r, "/login", http.StatusSeeOther)
		return
	}
	// Ambil ID user
	id, err := getUserIDFromSession(session)
	if err != nil {
		http.Error(w, "ID user tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	if id <= 0 {
		http.Error(w, "ID user tidak valid", http.StatusBadRequest)
		return
	}
	// Ambil data form
	namaLengkap := r.FormValue("nama_lengkap")
	email := r.FormValue("email")
	nomorHP := r.FormValue("nomor_hp")
	// Validasi
	if namaLengkap == "" {
		http.Error(w, "Nama lengkap tidak boleh kosong", http.StatusBadRequest)
		return
	}
	if email == "" {
		http.Error(w, "Email tidak boleh kosong", http.StatusBadRequest)
		return
	}
	if nomorHP == "" {
		http.Error(w, "Nomor HP tidak boleh kosong", http.StatusBadRequest)
		return
	}

	// Ambil user
	user, err := UserModel.GetByID(id)
	if err != nil {
		http.Error(w, "User tidak ditemukan: "+err.Error(), http.StatusNotFound)
		return
	}
	// update data
	user.NamaLengkap = namaLengkap
	user.Email = email
	user.NomerHP = nomorHP
	// simpan oauht database
	err = UserModel.Update(user)
	if err != nil {
		http.Error(w, "Gagal mengupdate profil: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// update session
	session.Values["nama_lengkap"] = user.NamaLengkap
	session.Values["email"] = user.Email
	session.Values["nomor_hp"] = user.NomerHP
	session.Values["picture"] = user.Picture
	session.Values["role"] = user.Role
	session.Values["provider"] = user.Provider

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "gagal Simpan session:"+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/profil", http.StatusSeeOther)
}
