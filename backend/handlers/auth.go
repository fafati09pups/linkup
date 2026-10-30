package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"linkup/database"

	"golang.org/x/crypto/bcrypt"
)

var secret = []byte(func() string {
	if v := os.Getenv("SECRET"); v != "" {
		return v
	}
	return "change-me-in-production"
}())

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func mac(p string) string {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(p))
	return hex.EncodeToString(m.Sum(nil))
}

func sign(uid int) string {
	p := fmt.Sprintf("%d.%d", uid, time.Now().Add(7*24*time.Hour).Unix())
	return p + "." + mac(p)
}

func userFrom(r *http.Request) (int, bool) {
	c, err := r.Cookie("token")
	if err != nil {
		return 0, false
	}
	i := strings.LastIndex(c.Value, ".")
	if i < 0 || !hmac.Equal([]byte(mac(c.Value[:i])), []byte(c.Value[i+1:])) {
		return 0, false
	}
	parts := strings.Split(c.Value[:i], ".")
	if len(parts) != 2 {
		return 0, false
	}
	uid, _ := strconv.Atoi(parts[0])
	exp, _ := strconv.ParseInt(parts[1], 10, 64)
	return uid, uid > 0 && time.Now().Unix() < exp
}

// Auth оборачивает обработчик проверкой входа.
func Auth(h func(http.ResponseWriter, *http.Request, int)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := userFrom(r)
		if !ok {
			fail(w, 401, "Войдите в аккаунт")
			return
		}
		h(w, r, uid)
	}
}

type creds struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var c creds
	json.NewDecoder(r.Body).Decode(&c)
	c.Username = strings.ToLower(strings.TrimSpace(c.Username))
	c.DisplayName = strings.TrimSpace(c.DisplayName)
	if len(c.Username) < 3 || len(c.Username) > 50 || len(c.Password) < 6 || c.DisplayName == "" {
		fail(w, 400, "Логин от 3 символов, пароль от 6, имя обязательно")
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
	_, err := database.DB.Exec(r.Context(),
		`INSERT INTO users (username, display_name, password_hash) VALUES ($1,$2,$3)`,
		c.Username, c.DisplayName, string(hash))
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			fail(w, 409, "Этот логин уже занят")
		} else {
			fail(w, 500, "Не удалось создать аккаунт")
		}
		return
	}
	writeJSON(w, 201, map[string]bool{"ok": true})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var c creds
	json.NewDecoder(r.Body).Decode(&c)
	var id int
	var hash string
	err := database.DB.QueryRow(r.Context(), `SELECT id, password_hash FROM users WHERE username=$1`,
		strings.ToLower(strings.TrimSpace(c.Username))).Scan(&id, &hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(c.Password)) != nil {
		fail(w, 401, "Неверный логин или пароль")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "token", Value: sign(id), Path: "/",
		HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 7 * 24 * 3600})
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "token", Path: "/", MaxAge: -1})
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func Me(w http.ResponseWriter, r *http.Request, uid int) {
	var u, d string
	database.DB.QueryRow(r.Context(), `SELECT username, display_name FROM users WHERE id=$1`, uid).Scan(&u, &d)
	writeJSON(w, 200, map[string]any{"id": uid, "username": u, "display_name": d})
}

func SearchUsers(w http.ResponseWriter, r *http.Request, uid int) {
	q := "%" + strings.TrimSpace(r.URL.Query().Get("q")) + "%"
	rows, err := database.DB.Query(r.Context(), `SELECT id, username, display_name FROM users
		WHERE id<>$1 AND (username ILIKE $2 OR display_name ILIKE $2) ORDER BY username LIMIT 20`, uid, q)
	if err != nil {
		fail(w, 500, "Ошибка поиска")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int
		var u, d string
		rows.Scan(&id, &u, &d)
		out = append(out, map[string]any{"id": id, "username": u, "display_name": d})
	}
	writeJSON(w, 200, out)
}

// Socket проверяет вход и передаёт соединение пакету ws.
func Socket(w http.ResponseWriter, r *http.Request) {
	uid, ok := userFrom(r)
	if !ok {
		fail(w, 401, "Войдите в аккаунт")
		return
	}
	wsServe(w, r, uid)
}
