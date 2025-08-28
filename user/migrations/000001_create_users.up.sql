CREATE TABLE IF NOT EXISTS users (
    id           BIGSERIAL PRIMARY KEY,
    telegram_id  BIGINT       NOT NULL UNIQUE,
    username     TEXT         NOT NULL,
    age          INTEGER      NOT NULL,
    gender       TEXT         NOT NULL,
    location     TEXT         NOT NULL,
    description  TEXT,
    photo_url    TEXT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_visible   BOOLEAN      NOT NULL DEFAULT TRUE
);
