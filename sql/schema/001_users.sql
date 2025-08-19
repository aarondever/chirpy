-- +goose Up
CREATE TABLE users(
    id UUID primary key,
    created_at timestamp not null,
    updated_at timestamp not null,
    email text not null unique
);

CREATE TABLE chirps(
    id UUID primary key,
    created_at timestamp not null,
    updated_at timestamp not null,
    body text not null unique,
    user_id UUID not null references users(id) on delete cascade
);

-- +goose Down
DROP TABLE chirps;
DROP TABLE users;