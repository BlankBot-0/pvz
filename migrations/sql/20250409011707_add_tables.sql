-- +goose Up
-- +goose StatementBegin
create table users
(
    id            uuid primary key,
    email         text unique,
    password_hash text,
    role          text
);

create table pvzs
(
    id                uuid primary key,
    registration_date timestamp default now(),
    city              text
);

create table receptions
(
    id     uuid primary key,
    date   timestamp default now(),
    status text,
    pvz_id uuid
);

create table products
(
    id           uuid primary key,
    date         timestamp default now(),
    type         text,

    reception_id uuid
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table products;
drop table receptions;
drop table pvzs;
drop table users;
-- +goose StatementEnd
