DROP TABLE IF EXISTS tasks;


CREATE TABLE tasks (
    id          SERIAL PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    completed   BOOLEAN   DEFAULT FALSE,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO tasks (title, description, completed)
VALUES
    ('Изучить GO', 'Пройти базовый курс', true),
    ('Написать REST API', 'Посмотреть видео и написать самому', false),
    ('Деплой приложения', 'Развернуть приложение на сервере', false)
