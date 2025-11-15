-- migrations/000002_create_unique_original_url.down.sql
-- Откат создания индекс на original_url
DROP INDEX IF EXISTS idx_shortened_urls_original_url_unique;
