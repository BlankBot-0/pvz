-- +goose Up
-- +goose StatementBegin
insert into cities (name)
values ('Москва'),
       ('Санкт-Петербург'),
       ('Казань');

insert into roles (name)
values ('employee'),
       ('moderator');

insert into product_types (name)
values ('электроника'),
       ('обувь'),
       ('одежда');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
truncate cities;
truncate roles;
truncate product_types;
-- +goose StatementEnd
