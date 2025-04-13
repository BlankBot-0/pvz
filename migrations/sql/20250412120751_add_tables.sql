-- +goose Up
-- +goose StatementBegin
create table users
(
    id            uuid primary key,
    email         text unique,
    password_hash text,
    role_id       int references roles(id)
);

create table pvzs
(
    id                uuid primary key,
    registration_date timestamp default now(),
    city_id           int references cities(id)
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
    type_id      bigint references product_types(id),

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
