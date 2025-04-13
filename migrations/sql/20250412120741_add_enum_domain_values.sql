-- +goose Up
-- +goose StatementBegin
create table cities
(
    id   serial primary key,
    name text unique
);

create table product_types
(
    id   bigserial primary key,
    name text unique
);

create table roles
(
    id   serial primary key,
    name text unique
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table product_types;
drop table cities;
drop table roles;
-- +goose StatementEnd
