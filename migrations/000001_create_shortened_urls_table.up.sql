-- migrations/000001_create_shortened_urls_table.up.sql
-- Создание таблицы сокращенных URL-ов
CREATE TABLE shortened_urls (
    id           SERIAL PRIMARY KEY,
    short_url    TEXT NOT NULL UNIQUE,
    original_url TEXT NOT NULL
);

-- Индекс на short_url для быстрого поиска при редиректе
CREATE INDEX idx_shortened_urls_short_url ON shortened_urls (short_url);