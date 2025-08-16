-- 01_create_tables.sql
CREATE TABLE orders (
                        id SERIAL PRIMARY KEY,
                        order_uid TEXT UNIQUE NOT NULL,
                        track_number TEXT NOT NULL,
                        entry TEXT,
                        locale TEXT,
                        internal_signature TEXT,
                        customer_id TEXT,
                        delivery_service TEXT,
                        shardkey TEXT,
                        sm_id INTEGER,
                        date_created TIMESTAMPTZ,
                        oof_shard TEXT
);

CREATE TABLE delivery (
                          id SERIAL PRIMARY KEY,
                          order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
                          name TEXT,
                          phone TEXT,
                          zip TEXT,
                          city TEXT,
                          address TEXT,
                          region TEXT,
                          email TEXT
);

CREATE TABLE payment (
                         id SERIAL PRIMARY KEY,
                         order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
                         transaction TEXT UNIQUE NOT NULL,
                         request_id TEXT,
                         currency TEXT,
                         provider TEXT,
                         amount NUMERIC,
                         payment_dt BIGINT,
                         bank TEXT,
                         delivery_cost NUMERIC,
                         goods_total NUMERIC,
                         custom_fee NUMERIC
);

CREATE TABLE items (
                       id SERIAL PRIMARY KEY,
                       order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
                       chrt_id BIGINT,
                       track_number TEXT,
                       price NUMERIC,
                       rid TEXT,
                       name TEXT,
                       sale INTEGER,
                       size TEXT,
                       total_price NUMERIC,
                       nm_id BIGINT,
                       brand TEXT,
                       status INTEGER
);
