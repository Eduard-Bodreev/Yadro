CREATE TABLE IF NOT EXISTS index_keywords (
    word TEXT NOT NULL,
    comic_id INTEGER NOT NULL,
    FOREIGN KEY (comic_id) REFERENCES comics(id),
    PRIMARY KEY (word, comic_id)
);