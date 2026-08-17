package main

import (
	"Auth/config"
	oauth "Auth/config/auth"
	"Auth/controller"
	oauthcontroller "Auth/controller/authhandler"
	"Auth/middleware"
	"log"
	"net/http"
	"text/template"

	"github.com/joho/godotenv"
)

var templates *template.Template

func main() {

	godotenv.Load()

	config.InitSessionStore()
	auth, err := oauth.NewAuthenticator()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	templates = template.Must(template.ParseGlob("views/*.html"))

	http.HandleFunc("/", controller.Index)
	http.HandleFunc("/loginauth", oauthcontroller.LoginHandler(auth))
	http.HandleFunc("/callback", oauthcontroller.CallbackHandler(auth))
	http.HandleFunc("/profil", oauthcontroller.UserHandler)
	http.HandleFunc("/logoutauth", oauthcontroller.LogoutHandler(auth))
	http.HandleFunc("/login", controller.Login)
	http.HandleFunc("/register", controller.Register)
	http.HandleFunc("/logout", controller.Logout)
	http.HandleFunc("/admin", middleware.Admin(controller.Admin))
	http.HandleFunc("/tambah", controller.Tambah)
	http.HandleFunc("/updatefoto", controller.UpdateFoto)
	http.HandleFunc("/updateprofil", controller.UpdateProfil)
	http.HandleFunc("/hapusproduk", controller.HapusProduk)
	http.HandleFunc("/admin/pesanan", controller.GetAllOrdersAdmin)
	http.HandleFunc("/admin/pesanan/status", controller.UpdateOrderStatusAdmin)
	http.HandleFunc("/pesanan", controller.CreateOrder)
	http.HandleFunc("/pesanan/saya", controller.GetUserOrders)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.Handle(
		"/uploads/",
		http.StripPrefix(
			"/uploads/",
			http.FileServer(http.Dir("uploads")),
		),
	)

	http.ListenAndServe(":8080", nil)
}
