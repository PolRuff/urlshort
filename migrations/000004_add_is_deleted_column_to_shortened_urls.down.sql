-- migrations/000004_add_is_deleted_column_to_shortened_urls.down.sql
-- Удаляем столбец is_deleted
ALTER TABLE shortened_urls DROP COLUMN is_deleted;
