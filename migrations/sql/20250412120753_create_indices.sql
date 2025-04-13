-- +goose Up
-- +goose StatementBegin
create index receptions_date_index on receptions (date);
create index products_date_index on products (date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index receptions_date_index;
drop index products_date_index;
-- +goose StatementEnd
