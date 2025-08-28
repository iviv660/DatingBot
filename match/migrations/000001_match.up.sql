CREATE TABLE IF NOT EXISTS matches (
                                       id BIGSERIAL PRIMARY KEY,
                                       from_user BIGINT NOT NULL,
                                       to_user   BIGINT NOT NULL,
                                       is_like   BOOLEAN NOT NULL,
                                       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (from_user, to_user)
    );

CREATE INDEX IF NOT EXISTS idx_matches_from_user ON matches(from_user);
CREATE INDEX IF NOT EXISTS idx_matches_to_user   ON matches(to_user);
