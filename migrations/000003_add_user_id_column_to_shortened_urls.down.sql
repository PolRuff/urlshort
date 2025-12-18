-- migrations/000003_add_user_id_column_to_shortened_urls.down.sql
-- Откат создания индекса на user_id
DROP INDEX IF EXISTS idx_shortened_urls_user_id;

-- Удаляем столбец user_id
ALTER TABLE shortened_urls DROP COLUMN user_id;
