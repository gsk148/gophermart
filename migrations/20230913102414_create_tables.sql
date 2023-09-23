-- +goose Up
CREATE TABLE IF NOT EXISTS users(
    id SERIAL PRIMARY KEY,
    login VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    balance MONEY NOT NULL DEFAULT 0,
    created_time TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS orders(
     id SERIAL PRIMARY KEY,
     number TEXT NOT NULL,
     user_id INT NOT NULL,
     uploaded_time TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
     operation_type TEXT DEFAULT 'ACCRUAL',
     status TEXT NOT NULL DEFAULT 'NEW',
     amount MONEY DEFAULT 0 NOT NULL,
     FOREIGN KEY(user_id) REFERENCES USERS(id)
);

-- +goose Down
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS orders;
