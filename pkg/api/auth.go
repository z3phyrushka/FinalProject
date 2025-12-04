package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type jwtPayload struct {
	Hash string `json:"hash"`
	Exp  int64  `json:"exp"`
}

func makeJWT(password string) (string, error) {

	header := `{"alg":"HS256","typ":"JWT"}`

	payload := jwtPayload{
		Hash: fmt.Sprintf("%x", sha256.Sum256([]byte(password))),
		Exp:  time.Now().Add(8 * time.Hour).Unix(),
	}

	pBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	h64 := base64.RawURLEncoding.EncodeToString([]byte(header))
	p64 := base64.RawURLEncoding.EncodeToString(pBytes)

	secret := []byte(password)

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(h64 + "." + p64))
	sign := h.Sum(nil)

	s64 := base64.RawURLEncoding.EncodeToString(sign)

	token := h64 + "." + p64 + "." + s64
	return token, nil
}

func validateJWT(token string, password string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}

	h64, p64, s64 := parts[0], parts[1], parts[2]

	pBytes, err := base64.RawURLEncoding.DecodeString(p64)
	if err != nil {
		return false
	}

	var payload jwtPayload
	if err := json.Unmarshal(pBytes, &payload); err != nil {
		return false
	}

	if time.Now().Unix() > payload.Exp {
		return false
	}

	expectHash := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	if payload.Hash != expectHash {
		return false
	}

	secret := []byte(password)

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(h64 + "." + p64))
	expectSign := h.Sum(nil)

	gotSign, err := base64.RawURLEncoding.DecodeString(s64)
	if err != nil {
		return false
	}

	return hmac.Equal(expectSign, gotSign)
}

func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		pass := os.Getenv("TODO_PASSWORD")
		if pass == "" {
			next(w, r)
			return
		}

		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		token := cookie.Value

		if !validateJWT(token, pass) {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
	}

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		writeJSON(w, map[string]string{"error": "invalid json"}, http.StatusBadRequest)
		return
	}

	pass := os.Getenv("TODO_PASSWORD")
	if pass == "" {
		writeJSON(w, map[string]string{"error": "password is not set on server"}, http.StatusUnauthorized)
		return
	}

	if req.Password != pass {
		writeJSON(w, map[string]string{"error": "Неверный пароль"}, http.StatusUnauthorized)
		return
	}

	token, err := makeJWT(pass)
	if err != nil {
		writeJSON(w, map[string]string{"error": "token generation failed"}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"token": token}, http.StatusOK)
}
