-- Initialize database for News Aggregator

-- Create database if not exists
-- This file is executed when the PostgreSQL container starts for the first time

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Ensure postgres superuser exists
DO
$do$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'postgres') THEN
      CREATE ROLE postgres WITH SUPERUSER LOGIN PASSWORD 'postgres';
   END IF;
END
$do$;

-- Create news table (matching application schema exactly)
CREATE TABLE IF NOT EXISTS news (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title TEXT NOT NULL,
    content TEXT,
    summary TEXT,
    url TEXT UNIQUE,
    image_url TEXT,
    author TEXT,
    source TEXT NOT NULL,
    category TEXT DEFAULT 'general',
    tags JSONB DEFAULT '[]',
    published_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    content_hash TEXT UNIQUE
);

-- Create indexes for better performance (matching application expectations)
CREATE INDEX IF NOT EXISTS idx_news_published_at ON news(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_news_source ON news(source);
CREATE INDEX IF NOT EXISTS idx_news_category ON news(category);
CREATE INDEX IF NOT EXISTS idx_news_content_hash ON news(content_hash);
CREATE INDEX IF NOT EXISTS idx_news_tags ON news USING GIN(tags);

-- Insert some sample data (optional)
-- You can uncomment and modify as needed

/*
-- Sample categories (will be created by the application)
INSERT INTO categories (name, description, color, icon) VALUES
('technology', 'Technology and innovation', '#3B82F6', '💻'),
('business', 'Business and finance', '#10B981', '💼'),
('sports', 'Sports and athletics', '#F59E0B', '⚽'),
('politics', 'Politics and government', '#EF4444', '🏛️'),
('health', 'Health and medicine', '#8B5CF6', '🏥'),
('science', 'Science and research', '#06B6D4', '🔬'),
('entertainment', 'Entertainment and media', '#F97316', '🎬'),
('world', 'World and international news', '#84CC16', '🌍')
ON CONFLICT (name) DO NOTHING;

-- Sample admin user (password: admin123)
INSERT INTO users (email, username, password_hash, first_name, last_name, is_admin, is_active) VALUES
('admin@newsaggregator.com', 'admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Admin', 'User', true, true)
ON CONFLICT (email) DO NOTHING;
*/
