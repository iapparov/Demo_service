CREATE TABLE IF NOT EXISTS orders (
    order_uid TEXT PRIMARY KEY,
    track_number TEXT NOT NULL,
    entry TEXT NOT NULL,
    locale TEXT,
    internal_signature TEXT,
    customer_id TEXT,
    delivery_service TEXT,
    shardkey SMALLINT,
    sm_id INTEGER,
    date_created TIMESTAMP WITH TIME ZONE,
    oof_shard SMALLINT
);

CREATE TABLE IF NOT EXISTS deliveries (
    id SERIAL PRIMARY KEY,
    order_uid TEXT REFERENCES orders(order_uid) ON DELETE CASCADE,
    name TEXT,
    phone TEXT,
    zip TEXT,
    city TEXT,
    address TEXT,
    region TEXT,
    email TEXT
);

CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    order_uid TEXT REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction TEXT,
    request_id TEXT,
    currency TEXT,
    provider TEXT,
    amount INTEGER,
    payment_dt TIMESTAMP WITH TIME ZONE,
    bank TEXT,
    delivery_cost INTEGER,
    goods_total INTEGER,
    custom_fee INTEGER
);

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    order_uid TEXT REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id BIGINT,
    track_number TEXT,
    price INTEGER,
    rid TEXT,
    name TEXT,
    sale SMALLINT,
    size TEXT,
    total_price INTEGER,
    nm_id BIGINT,
    brand TEXT,
    status SMALLINT
);

