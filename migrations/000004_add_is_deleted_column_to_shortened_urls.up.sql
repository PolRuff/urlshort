-- migrations/000004_add_is_deleted_column_to_shortened_urls.up.sql
-- Добавляем столбец is_deleted
ALTER TABLE shortened_urls ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE;
