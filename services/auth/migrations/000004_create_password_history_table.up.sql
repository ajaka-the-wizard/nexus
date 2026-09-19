CREATE TABLE IF NOT EXISTS password_history(
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    retired_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (created_at < retired_at)
);

CREATE INDEX idx_password_history_user_id ON password_history(user_id);
CREATE INDEX idx_password_history_created_at ON password_history(created_at);
CREATE INDEX idx_password_history_retired_at ON password_history(retired_at);
