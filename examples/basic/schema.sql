CREATE TYPE user_status AS ENUM ('active', 'inactive', 'deleted');

CREATE TABLE users (
    id            bigserial PRIMARY KEY,
    email         text NOT NULL,
    name          text NOT NULL,
    status        user_status NOT NULL DEFAULT 'active',
    tags          text[],
    meta          jsonb,
    balance       numeric(12, 2),
    external_id   uuid,
    last_login_at timestamptz,
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE posts (
    id         bigserial PRIMARY KEY,
    user_id    bigint NOT NULL REFERENCES users(id),
    title      text NOT NULL,
    published  boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- ignored by generate.ExcludeTables
CREATE TABLE _skip_me (
    id bigint PRIMARY KEY
);
