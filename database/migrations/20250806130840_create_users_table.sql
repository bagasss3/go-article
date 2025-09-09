-- +goose Up
-- +goose StatementBegin
CREATE TABLE authors (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

-- Create articles table
CREATE TABLE articles (
    id TEXT PRIMARY KEY,
    author_id TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_author FOREIGN KEY (author_id) REFERENCES authors(id) ON DELETE CASCADE
);

CREATE INDEX idx_articles_author_id ON articles(author_id);
CREATE INDEX idx_articles_created_at ON articles(created_at DESC);
CREATE INDEX idx_articles_title_search ON articles USING GIN (to_tsvector('simple', title));
CREATE INDEX idx_articles_body_search ON articles USING GIN (to_tsvector('simple', body));
CREATE INDEX idx_articles_title_body_search ON articles USING GIN (to_tsvector('simple', title || ' ' || body));
CREATE INDEX idx_authors_name ON authors(name);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_authors_name;
DROP INDEX IF EXISTS idx_articles_title_body_search;
DROP INDEX IF EXISTS idx_articles_body_search;
DROP INDEX IF EXISTS idx_articles_title_search;
DROP INDEX IF EXISTS idx_articles_created_at;
DROP INDEX IF EXISTS idx_articles_author_id;
DROP TABLE IF EXISTS articles;
DROP TABLE IF EXISTS authors;
-- +goose StatementEnd
