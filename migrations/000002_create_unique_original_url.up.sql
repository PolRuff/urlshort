-- migrations/000002_create_unique_original_url.up.sql
-- Индекс на original_url для избавления от дубликатов
CREATE UNIQUE INDEX idx_shortened_urls_original_url_unique ON shortened_urls (original_url);
