package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"linkup/database"
	"linkup/models"
	"linkup/ws"
)

var wsServe = ws.Serve

func ListChats(w http.ResponseWriter, r *http.Request, uid int) {
	rows, err := database.DB.Query(r.Context(), `
		SELECT c.id, u.id, u.username, u.display_name, COALESCE(m.content,''), m.created_at
		FROM chats c
		JOIN chat_members me ON me.chat_id=c.id AND me.user_id=$1
		JOIN chat_members o  ON o.chat_id=c.id AND o.user_id<>me.user_id
		JOIN users u ON u.id=o.user_id
		LEFT JOIN LATERAL (SELECT content, created_at FROM messages
			WHERE chat_id=c.id ORDER BY created_at DESC LIMIT 1) m ON true
		WHERE c.chat_type='direct'
		ORDER BY COALESCE(m.created_at, c.created_at) DESC`, uid)
	if err != nil {
		fail(w, 500, "Не удалось загрузить чаты")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, uid2 int
		var un, dn, last string
		var at *time.Time
		rows.Scan(&id, &uid2, &un, &dn, &last, &at)
		out = append(out, map[string]any{"id": id, "user_id": uid2, "username": un, "name": dn, "last": last, "at": at})
	}
	writeJSON(w, 200, out)
}

// OpenChat находит личный чат или создаёт его в транзакции.
func OpenChat(w http.ResponseWriter, r *http.Request, uid int) {
	var in struct {
		UserID int `json:"user_id"`
	}
	json.NewDecoder(r.Body).Decode(&in)
	if in.UserID == 0 || in.UserID == uid {
		fail(w, 400, "Неверный собеседник")
		return
	}
	ctx, db := r.Context(), database.DB
	var chatID int
	err := db.QueryRow(ctx, `SELECT c.id FROM chats c
		JOIN chat_members a ON a.chat_id=c.id AND a.user_id=$1
		JOIN chat_members b ON b.chat_id=c.id AND b.user_id=$2
		WHERE c.chat_type='direct' LIMIT 1`, uid, in.UserID).Scan(&chatID)
	if err != nil {
		tx, err := db.Begin(ctx)
		if err != nil {
			fail(w, 500, "Ошибка базы данных")
			return
		}
		defer tx.Rollback(ctx)
		if err = tx.QueryRow(ctx, `INSERT INTO chats (chat_type) VALUES ('direct') RETURNING id`).Scan(&chatID); err == nil {
			_, err = tx.Exec(ctx, `INSERT INTO chat_members (chat_id, user_id) VALUES ($1,$2),($1,$3)`, chatID, uid, in.UserID)
		}
		if err == nil {
			err = tx.Commit(ctx)
		}
		if err != nil {
			fail(w, 400, "Не удалось создать чат")
			return
		}
	}
	writeJSON(w, 200, map[string]int{"id": chatID})
}

func History(w http.ResponseWriter, r *http.Request, uid int) {
	chat, _ := strconv.Atoi(r.PathValue("id"))
	if !database.IsMember(r.Context(), chat, uid) {
		fail(w, 403, "Нет доступа к чату")
		return
	}
	rows, err := database.DB.Query(r.Context(), `SELECT m.id, m.chat_id, m.sender_id, u.username, m.content, m.created_at
		FROM messages m JOIN users u ON u.id=m.sender_id
		WHERE m.chat_id=$1 ORDER BY m.created_at ASC, m.id ASC`, chat)
	if err != nil {
		fail(w, 500, "Не удалось загрузить сообщения")
		return
	}
	defer rows.Close()
	out := []models.Msg{}
	for rows.Next() {
		var m models.Msg
		rows.Scan(&m.ID, &m.ChatID, &m.SenderID, &m.Username, &m.Content, &m.At)
		out = append(out, m)
	}
	writeJSON(w, 200, out)
}

func Send(w http.ResponseWriter, r *http.Request, uid int) {
	chat, _ := strconv.Atoi(r.PathValue("id"))
	ctx := r.Context()
	if !database.IsMember(ctx, chat, uid) {
		fail(w, 403, "Нет доступа к чату")
		return
	}
	var in struct {
		Content string `json:"content"`
	}
	json.NewDecoder(r.Body).Decode(&in)
	in.Content = strings.TrimSpace(in.Content)
	if in.Content == "" || len(in.Content) > 4000 {
		fail(w, 400, "Сообщение пустое или слишком длинное")
		return
	}
	var blocked bool
	database.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM user_blocks b
		JOIN chat_members m ON m.chat_id=$1 AND m.user_id=b.blocker_id
		WHERE b.blocked_id=$2)`, chat, uid).Scan(&blocked)
	if blocked {
		fail(w, 403, "Пользователь вас заблокировал")
		return
	}
	m := models.Msg{ChatID: chat, SenderID: uid, Content: in.Content}
	err := database.DB.QueryRow(ctx, `WITH ins AS (
			INSERT INTO messages (chat_id, sender_id, content) VALUES ($1,$2,$3) RETURNING id, created_at)
		SELECT ins.id, ins.created_at, u.username FROM ins, users u WHERE u.id=$2`,
		chat, uid, in.Content).Scan(&m.ID, &m.At, &m.Username)
	if err != nil {
		fail(w, 500, "Не удалось отправить сообщение")
		return
	}
	data, _ := json.Marshal(map[string]any{"type": "message", "chat_id": chat, "message": m})
	ws.Send(database.Members(ctx, chat), data)
	writeJSON(w, 201, m)
}
