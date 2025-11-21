-- migrations/000001_create_shortened_urls_table.down.sql
-- Откат создания таблицы сокращенных URL-ов
DROP INDEX IF EXISTS idx_shortened_urls_short_url;
DROP TABLE IF EXISTS shortened_urls;
