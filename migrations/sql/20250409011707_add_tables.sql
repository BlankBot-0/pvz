-- +goose Up
-- +goose StatementBegin
create table users
(
    id uuid primary key,
    email string,
    password_hash string,
    role string
);

create table pvzs
(
    id uuid primary key,
    registration_date timestamp default now(),
    city string
);

create table receptions
(
    id uuid primary key,
    date timestamp default now(),
    status string,
    pvz_id uuid,
);

create table products
(
    id uuid primary key,
    date timestamp default now(),
    type string,

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
