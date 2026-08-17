package middleware

import (
	"Auth/config"
	"net/http"
)

func Admin(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		session, err := config.Store.Get(r, config.SESSION_ID)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		// Belum login
		if session.Values["LoggedIn"] != true {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		// Cek role
		role, ok := session.Values["role"].(string)
		if !ok || role != "admin" {
			http.Error(w, "Akses ditolak. Hanya admin yang boleh mengakses halaman ini.", http.StatusForbidden)
			return
		}
		// Kalau lolos semua pengecekan
		next.ServeHTTP(w, r)
	}
}
