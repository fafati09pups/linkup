package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect(url string) error {
	p, err := pgxpool.New(context.Background(), url)
	DB = p
	return err
}

// IsMember проверяет, участвует ли пользователь в чате.
func IsMember(ctx context.Context, chat, uid int) bool {
	var ok bool
	DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM chat_members WHERE chat_id=$1 AND user_id=$2)`, chat, uid).Scan(&ok)
	return ok
}

// Members возвращает ID всех участников чата.
func Members(ctx context.Context, chat int) []int {
	rows, err := DB.Query(ctx, `SELECT user_id FROM chat_members WHERE chat_id=$1`, chat)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		rows.Scan(&id)
		ids = append(ids, id)
	}
	return ids
}
