-- +goose Up
-- +goose StatementBegin
insert into cities (name)
values ('Москва'),
       ('Санкт-Петербург'),
       ('Казань');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- truncate cities;
-- +goose StatementEnd
