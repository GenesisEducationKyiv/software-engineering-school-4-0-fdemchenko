CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    email TEXT UNIQUE NOT NULL
);


