-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS verification (
  `id` INTEGER NOT NULL PRIMARY KEY,
  `code` VARCHAR(255) NOT NULL,
  `user` INTEGER NOT NULL,
  `expiration` TIMESTAMP NOT NULL,
  FOREIGN KEY (user) REFERENCES users(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE verification;
-- +goose StatementEnd