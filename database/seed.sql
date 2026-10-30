-- Тестовые данные. Запускать один раз на пустой базе после schema.sql:
-- psql linkup -f database/seed.sql
-- Пароль всех тестовых пользователей: 123456

CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO users (username, display_name, password_hash) VALUES
    ('anna',  'Анна',  crypt('123456', gen_salt('bf', 10))),
    ('boris', 'Борис', crypt('123456', gen_salt('bf', 10))),
    ('clara', 'Клара', crypt('123456', gen_salt('bf', 10)));

INSERT INTO chats (chat_type) VALUES ('direct'), ('direct');

INSERT INTO chat_members (chat_id, user_id) VALUES (1, 1), (1, 2), (2, 1), (2, 3);

INSERT INTO messages (chat_id, sender_id, content) VALUES
    (1, 1, 'Привет, Борис!'),
    (1, 2, 'Привет! Как дела?'),
    (1, 1, 'Отлично, делаю проект LINKUP'),
    (2, 3, 'Анна, ты придёшь завтра?');

INSERT INTO message_reads (message_id, user_id) VALUES (1, 2), (2, 1);

INSERT INTO user_blocks (blocker_id, blocked_id) VALUES (3, 2);
