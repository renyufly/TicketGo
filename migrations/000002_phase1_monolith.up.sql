CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'customer' CHECK (role IN ('customer', 'admin')),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX users_email_unique ON users (LOWER(email));

CREATE TABLE items (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL CHECK (LENGTH(BTRIM(name)) BETWEEN 1 AND 200),
    description TEXT NOT NULL DEFAULT '',
    price_cents BIGINT NOT NULL CHECK (price_cents BETWEEN 1 AND 1000000000000),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE activities (
    id BIGSERIAL PRIMARY KEY,
    item_id BIGINT NOT NULL REFERENCES items(id) ON DELETE RESTRICT,
    name TEXT NOT NULL CHECK (LENGTH(BTRIM(name)) BETWEEN 1 AND 200),
    price_cents BIGINT NOT NULL CHECK (price_cents BETWEEN 1 AND 1000000000000),
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'ended', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (ends_at > starts_at)
);
CREATE INDEX activities_item_id_idx ON activities (item_id);
CREATE INDEX activities_status_starts_at_idx ON activities (status, starts_at);

CREATE TABLE inventories (
    id BIGSERIAL PRIMARY KEY,
    activity_id BIGINT NOT NULL UNIQUE REFERENCES activities(id) ON DELETE RESTRICT,
    total BIGINT NOT NULL CHECK (total > 0),
    available BIGINT NOT NULL CHECK (available >= 0),
    sold BIGINT NOT NULL DEFAULT 0 CHECK (sold >= 0),
    version BIGINT NOT NULL DEFAULT 0 CHECK (version >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (sold + available = total)
);

CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    order_no TEXT NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    activity_id BIGINT NOT NULL REFERENCES activities(id) ON DELETE RESTRICT,
    quantity BIGINT NOT NULL CHECK (quantity > 0),
    unit_price_cents BIGINT NOT NULL CHECK (unit_price_cents > 0),
    total_price_cents BIGINT NOT NULL CHECK (total_price_cents > 0),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cancelled_at TIMESTAMPTZ
);
CREATE INDEX orders_user_created_at_idx ON orders (user_id, created_at DESC, id DESC);
CREATE INDEX orders_activity_id_idx ON orders (activity_id);

CREATE TABLE seckill_records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    activity_id BIGINT NOT NULL REFERENCES activities(id) ON DELETE RESTRICT,
    order_id BIGINT NOT NULL UNIQUE REFERENCES orders(id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'succeeded' CHECK (status IN ('succeeded', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, activity_id)
);
