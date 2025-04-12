-- +goose Up
-- +goose StatementBegin
create table cities
(
    name text unique
);

create table product_types
(
    name text unique
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table product_types;
drop table cities;
-- +goose StatementEnd
