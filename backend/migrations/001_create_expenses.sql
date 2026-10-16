-- +goose Up
CREATE TABLE IF NOT EXISTS expenses (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL CHECK (user_id > 0),
    amount INTEGER NOT NULL CHECK (amount > 0),
    category TEXT NOT NULL,
    spent_at TEXT NOT NULL,
    comment TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- +goose Down
DROP TABLE expenses;
