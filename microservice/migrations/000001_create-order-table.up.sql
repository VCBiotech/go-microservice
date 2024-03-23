CREATE TABLE users (
    id serial PRIMARY KEY,
    email varchar(255) UNIQUE NOT NULL,
    password_hash varchar(255), -- For local authentication, if used
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE oauth_providers (
    id serial PRIMARY KEY,
    name varchar(50) UNIQUE NOT NULL
);

CREATE TABLE user_oauth (
    id serial PRIMARY KEY,
    user_id integer NOT NULL REFERENCES users (id),
    provider_id integer NOT NULL REFERENCES oauth_providers (id),
    provider_user_id varchar(255) NOT NULL,
    access_token varchar(255),
    refresh_token varchar(255),
    token_expires timestamp with time zone,
    UNIQUE (user_id, provider_id),
    UNIQUE (provider_id, provider_user_id)
);

CREATE TABLE sessions (
    id serial PRIMARY KEY,
    user_id integer NOT NULL REFERENCES users (id),
    session_token varchar(255) UNIQUE NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE orders (
    order_id serial PRIMARY KEY,
    customer_id int NOT NULL,
    order_date date NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL
);
