package ws

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	mu    sync.Mutex
	conns = map[int]map[*websocket.Conn]bool{}
	up    = websocket.Upgrader{} // по умолчанию проверяет Origin
)

// Send отправляет данные всем открытым соединениям указанных пользователей.
func Send(userIDs []int, data []byte) {
	mu.Lock()
	defer mu.Unlock()
	for _, id := range userIDs {
		for c := range conns[id] {
			c.WriteMessage(websocket.TextMessage, data)
		}
	}
}

// Serve держит соединение пользователя, пока оно не закроется.
func Serve(w http.ResponseWriter, r *http.Request, uid int) {
	c, err := up.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	mu.Lock()
	if conns[uid] == nil {
		conns[uid] = map[*websocket.Conn]bool{}
	}
	conns[uid][c] = true
	mu.Unlock()
	defer func() {
		mu.Lock()
		delete(conns[uid], c)
		mu.Unlock()
		c.Close()
	}()
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			return
		}
	}
}
