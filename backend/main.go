package main

import (
	"log"
	"net/http"
	"os"

	"linkup/database"
	"linkup/handlers"
)

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	if err := database.Connect(getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/linkup")); err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/register", handlers.Register)
	mux.HandleFunc("POST /api/login", handlers.Login)
	mux.HandleFunc("POST /api/logout", handlers.Logout)
	mux.HandleFunc("GET /api/me", handlers.Auth(handlers.Me))
	mux.HandleFunc("GET /api/users", handlers.Auth(handlers.SearchUsers))
	mux.HandleFunc("GET /api/chats", handlers.Auth(handlers.ListChats))
	mux.HandleFunc("POST /api/chats", handlers.Auth(handlers.OpenChat))
	mux.HandleFunc("GET /api/chats/{id}/messages", handlers.Auth(handlers.History))
	mux.HandleFunc("POST /api/chats/{id}/messages", handlers.Auth(handlers.Send))
	mux.HandleFunc("GET /ws", handlers.Socket)
	mux.Handle("/", http.FileServer(http.Dir(getenv("FRONTEND", "../frontend"))))

	addr := getenv("ADDR", ":8080")
	log.Println("LINKUP запущен на http://localhost" + addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
