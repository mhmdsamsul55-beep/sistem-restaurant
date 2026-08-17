package authhandler

import (
	"Auth/config"
	"Auth/config/auth"
	"Auth/entities"
	usermodel "Auth/models"

	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/gob"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/auth0/go-auth0/v2/authentication/oauth"
	"github.com/gorilla/sessions"
)

func init() {
	gob.Register(map[string]interface{}{})
}

// =====================================================
// INIT SESSION STORE
// =====================================================

func InitSessionStore() {

	secret := os.Getenv("SESSION_SECRET")

	if secret == "" {
		secret = "eewffbwbfebbfhwf"
	}

	config.Store = sessions.NewCookieStore(
		[]byte(secret),
	)

	config.Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
}

// =====================================================
// HOME
// =====================================================

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	session, _ := config.Store.Get(
		r,
		"auth-session",
	)

	if session.Values["profile"] != nil {

		http.Redirect(
			w,
			r,
			"/",
			http.StatusSeeOther,
		)

		return
	}

	temp, err := template.ParseFiles(
		"views/index.html",
	)

	if err != nil {

		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	temp.Execute(w, nil)
}

// =====================================================
// LOGIN AUTH0
// =====================================================

func LoginHandler(auth *auth.Authenticator) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		state, err := generateRandomState()

		if err != nil {

			http.Error(
				w,
				"Internal error",
				http.StatusInternalServerError,
			)

			return
		}

		session, _ := config.Store.Get(
			r,
			"auth-session",
		)

		session.Values["state"] = state

		if err := session.Save(r, w); err != nil {

			http.Error(
				w,
				"Internal error",
				http.StatusInternalServerError,
			)

			return
		}

		http.Redirect(
			w,
			r,
			auth.AuthorizationURL(state),
			http.StatusTemporaryRedirect,
		)
	}
}

// =====================================================
// CALLBACK AUTH0
// =====================================================

func CallbackHandler(auth *auth.Authenticator) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// =================================================
		// 1. AMBIL AUTH SESSION
		// =================================================

		session, err := config.Store.Get(
			r,
			"auth-session",
		)

		if err != nil {

			http.Error(
				w,
				"Gagal mengambil auth session: "+err.Error(),
				http.StatusInternalServerError,
			)

			return
		}

		// =================================================
		// 2. VALIDASI STATE
		// =================================================

		state := r.URL.Query().Get("state")

		savedState, ok := session.Values["state"].(string)

		if !ok || savedState == "" {

			http.Error(
				w,
				"State session tidak ditemukan",
				http.StatusBadRequest,
			)

			return
		}

		if state != savedState {

			http.Error(
				w,
				"Invalid state parameter",
				http.StatusBadRequest,
			)

			return
		}

		// =================================================
		// 3. AMBIL AUTHORIZATION CODE
		// =================================================

		code := r.URL.Query().Get("code")

		if code == "" {

			http.Error(
				w,
				"Authorization code tidak ditemukan",
				http.StatusBadRequest,
			)

			return
		}

		// =================================================
		// 4. TUKAR CODE MENJADI TOKEN
		// =================================================

		tokenSet, err := auth.OAuth.LoginWithAuthCode(
			r.Context(),
			oauth.LoginWithAuthCodeRequest{
				Code:        code,
				RedirectURI: auth.CallbackURL,
			},
			oauth.IDTokenValidationOptions{},
		)

		if err != nil {

			http.Error(
				w,
				"Failed to exchange authorization code for token: "+err.Error(),
				http.StatusUnauthorized,
			)

			return
		}

		// =================================================
		// 5. AMBIL DATA USER DARI AUTH0 / GOOGLE
		// =================================================

		userInfo, err := auth.UserInfo(
			r.Context(),
			tokenSet.AccessToken,
		)

		if err != nil {

			http.Error(
				w,
				"Gagal mengambil user info: "+err.Error(),
				http.StatusInternalServerError,
			)

			return
		}

		// =================================================
		// 6. SIMPAN DATA AUTH0 KE AUTH SESSION
		// =================================================

		session.Values["access_token"] =
			tokenSet.AccessToken

		session.Values["profile"] = map[string]interface{}{
			"nama_lengkap": userInfo.Name,
			"email":        userInfo.Email,
			"picture":      userInfo.Picture,
			"provider":     "google",
		}

		UserModel := usermodel.NewUserModel()

		user, err := UserModel.FindByEmail(
			userInfo.Email,
		)

		if err == sql.ErrNoRows {

			user := entities.User{
				NamaLengkap: userInfo.Name,
				Email:       userInfo.Email,
				NomerHP:     "",
				Provider:    "google",
				Role:        "user",
				Picture:     userInfo.Picture,
			}

			err = UserModel.Create(&user)

			if err != nil {

				http.Error(
					w,
					"Gagal menyimpan user ke database: "+err.Error(),
					http.StatusInternalServerError,
				)

				return
			}

			user, err = UserModel.FindByEmail(
				user.Email,
			)

			if err != nil {

				http.Error(
					w,
					"Gagal mengambil data user setelah dibuat: "+err.Error(),
					http.StatusInternalServerError,
				)

				return
			}
		}

		if err != nil {

			http.Error(w,"Gagal mencari user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// =================================================
		// USER SUDAH ADA
		// JANGAN CREATE LAGI
		// =================================================

		if user.NamaLengkap == "" {

			user.NamaLengkap =
				userInfo.Name
		}

		if user.Picture == "" {

			user.Picture =
				userInfo.Picture
		}

		if user.Provider == "" {

			user.Provider =
				"google"
		}
		delete(
			session.Values,
			"state",
		)

		if err := session.Save(r, w); err != nil {

			http.Error(
				w,
				"Gagal menyimpan auth session: "+err.Error(),
				http.StatusInternalServerError,
			)

			return
		}
		mainSession, err := config.Store.Get(
			r,
			config.SESSION_ID,
		)

		if err != nil {

			http.Error(
				w,
				"Gagal mengambil main session: "+err.Error(),
				http.StatusInternalServerError,
			)

			return
		}

		// =================================================
		// 10. SIMPAN USER KE MAIN SESSION
		// =================================================

		mainSession.Values["LoggedIn"] =
			true

		// ID USER DARI DATABASE
		mainSession.Values["id"] =
			strconv.FormatInt(
				user.ID,
				10,
			)

		mainSession.Values["nama_lengkap"] =
			user.NamaLengkap

		mainSession.Values["email"] =
			user.Email

		mainSession.Values["picture"] =
			user.Picture

		mainSession.Values["provider"] =
			user.Provider

		mainSession.Values["nomor_hp"] =
			user.NomerHP

		mainSession.Values["role"] =
			user.Role

		// =================================================
		// 11. SIMPAN MAIN SESSION
		// =================================================

		if err := mainSession.Save(r, w); err != nil {

			http.Error(
				w,
				"Gagal menyimpan main session: "+err.Error(),
				http.StatusInternalServerError,
			)

			return
		}

		// =================================================
		// 12. REDIRECT KE HOME
		// =================================================

		http.Redirect(
			w,
			r,
			"/",
			http.StatusSeeOther,
		)
	}
}

// =====================================================
// USER / PROFILE
// =====================================================

func UserHandler(w http.ResponseWriter, r *http.Request) {

	session, err := config.Store.Get(
		r,
		config.SESSION_ID,
	)

	if err != nil {

		http.Error(
			w,
			"Gagal mengambil session",
			http.StatusInternalServerError,
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

	data := map[string]interface{}{

		"id": session.Values["id"],

		"nama_lengkap": session.Values["nama_lengkap"],

		"email": session.Values["email"],

		"nomor_hp": session.Values["nomor_hp"],

		"provider": session.Values["provider"],

		"picture": session.Values["picture"],

		"role": session.Values["role"],
	}

	temp, err := template.ParseFiles(
		"views/profil.html",
	)

	if err != nil {

		http.Error(
			w,
			"Gagal membuka profil: "+err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	err = temp.Execute(
		w,
		data,
	)

	if err != nil {

		http.Error(
			w,
			"Gagal menampilkan profil: "+err.Error(),
			http.StatusInternalServerError,
		)

		return
	}
}

// =====================================================
// LOGOUT
// =====================================================

func LogoutHandler(auth *auth.Authenticator) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// ---------------------------------------------
		// HAPUS AUTH SESSION
		// ---------------------------------------------

		session, _ := config.Store.Get(
			r,
			"auth-session",
		)

		session.Options.MaxAge = -1

		session.Save(
			r,
			w,
		)

		// ---------------------------------------------
		// HAPUS MAIN SESSION
		// ---------------------------------------------

		mainSession, _ := config.Store.Get(
			r,
			config.SESSION_ID,
		)

		mainSession.Options.MaxAge = -1

		mainSession.Save(
			r,
			w,
		)

		// ---------------------------------------------
		// LOGOUT AUTH0
		// ---------------------------------------------

		logoutURL, _ := url.Parse(
			"https://" + auth.Domain + "/v2/logout",
		)

		scheme := "http"

		if r.TLS != nil {
			scheme = "https"
		}

		returnTo, _ := url.Parse(
			scheme + "://" + r.Host,
		)

		params := url.Values{}

		params.Add(
			"returnTo",
			returnTo.String(),
		)

		params.Add(
			"client_id",
			auth.ClientID,
		)

		logoutURL.RawQuery =
			params.Encode()

		http.Redirect(
			w,
			r,
			logoutURL.String(),
			http.StatusTemporaryRedirect,
		)
	}
}

// =====================================================
// GENERATE RANDOM STATE
// =====================================================

func generateRandomState() (string, error) {

	b := make(
		[]byte,
		32,
	)

	_, err := rand.Read(b)

	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
