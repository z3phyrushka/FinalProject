package api

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var serverPassword string

func init() {
	serverPassword = os.Getenv("TODO_PASSWORD")
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	if serverPassword == "" {
		writeJSON(w, map[string]string{"error": "Password not set on server"}, http.StatusBadRequest)
		return
	}

	var payload struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, map[string]string{"error": "Invalid request"}, http.StatusBadRequest)
		return
	}

	if payload.Password != serverPassword {
		writeJSON(w, map[string]string{"error": "Неверный пароль"}, http.StatusUnauthorized)
		return
	}

	claims := jwt.MapClaims{
		"hash": payload.Password,
		"exp":  time.Now().Add(8 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(serverPassword))
	if err != nil {
		writeJSON(w, map[string]string{"error": "Token generation failed"}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"token": tokenStr}, http.StatusOK)
}

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if serverPassword != "" {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (any, error) {
				return []byte(serverPassword), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
		}

		next(w, r)
	}
}
