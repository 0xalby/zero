-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS users (
  `id` INTEGER NOT NULL PRIMARY KEY,
  `customer` VARCHAR(255) NOT NULL UNIQUE, -- Stripe customer id
  `username` VARCHAR(255) NOT NULL UNIQUE,
  `email` VARCHAR(255) NOT NULL UNIQUE,
  `password` VARCHAR(255) NOT NULL,
  `verified` BOOLEAN NOT NULL DEFAULT 0, -- Email verification
  `updated` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE users;
-- +goose StatementEnd
