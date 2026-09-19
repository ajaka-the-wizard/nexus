CREATE TYPE provider_enum AS ENUM ('google','github','custom');

CREATE TABLE IF NOT EXISTS providers(
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider provider_enum NOT NULL,
    provider_sub TEXT NOT NULL,
    verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, provider),
    UNIQUE (provider, provider_sub)
);

CREATE INDEX idx_providers_created_at ON providers(created_at);
CREATE INDEX idx_providers_updated_at ON providers(updated_at);