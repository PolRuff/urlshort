-- migrations/000003_add_user_id_column_to_shortened_urls.up.sql
-- Добавляем столбец user_id
ALTER TABLE shortened_urls ADD COLUMN user_id BIGINT NOT NULL DEFAULT 0;

-- Создаём индекс для поиска всех URL по user_id
CREATE INDEX idx_shortened_urls_user_id ON shortened_urls (user_id);
