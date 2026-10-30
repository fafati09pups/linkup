-- Запросы для защиты. Выполняйте по одному после schema.sql и seed.sql.

-- INSERT: новое сообщение
INSERT INTO messages (chat_id, sender_id, content) VALUES (1, 2, 'Тестовое сообщение');

-- SELECT + JOIN + ORDER BY: история переписки
SELECT m.id, u.username, m.content, m.created_at
FROM messages m
JOIN users u ON u.id = m.sender_id
WHERE m.chat_id = 1
ORDER BY m.created_at ASC;

-- JOIN трёх таблиц: участники каждого чата
SELECT c.id AS chat_id, c.chat_type, u.username
FROM chats c
JOIN chat_members cm ON cm.chat_id = c.id
JOIN users u ON u.id = cm.user_id
ORDER BY c.id, u.username;

-- GROUP BY + COUNT: сколько сообщений отправил каждый пользователь
SELECT u.username, COUNT(m.id) AS messages_count
FROM users u
LEFT JOIN messages m ON m.sender_id = u.id
GROUP BY u.username
ORDER BY messages_count DESC;

-- Непрочитанные сообщения пользователя 1 (message_reads)
SELECT m.chat_id, COUNT(*) AS unread
FROM messages m
JOIN chat_members cm ON cm.chat_id = m.chat_id AND cm.user_id = 1
LEFT JOIN message_reads r ON r.message_id = m.id AND r.user_id = 1
WHERE m.sender_id <> 1 AND r.message_id IS NULL
GROUP BY m.chat_id;

-- UPDATE: редактирование сообщения
UPDATE messages SET content = 'Исправленный текст', edited_at = now() WHERE id = 1;

-- DELETE: удаление сообщения (связанные записи message_reads удалятся каскадом)
DELETE FROM messages WHERE id = 4;

-- Транзакция: создание чата вместе с участниками
BEGIN;
INSERT INTO chats (chat_type) VALUES ('direct') RETURNING id;
-- подставьте полученный id вместо 3
INSERT INTO chat_members (chat_id, user_id) VALUES (3, 2), (3, 3);
COMMIT;   -- или ROLLBACK, чтобы отменить обе вставки

-- Проверка ограничений (каждый запрос должен вернуть ошибку)
INSERT INTO users (username, display_name, password_hash) VALUES ('anna', 'Дубль', 'x');  -- UNIQUE
INSERT INTO messages (chat_id, sender_id, content) VALUES (1, 1, '   ');                  -- CHECK
INSERT INTO user_blocks (blocker_id, blocked_id) VALUES (1, 1);                           -- CHECK

-- Индексы: план выполнения запроса истории
EXPLAIN ANALYZE SELECT * FROM messages WHERE chat_id = 1 ORDER BY created_at;
