-- +goose Up
-- +goose StatementBegin
create table if not exists encryptors
(
    id          bigserial primary key,
    engine      varchar(255) not null,
    encryptor_type  varchar(255) not null,
    additional jsonb,
    created_at  timestamp default now() not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists secrets;
drop table if exists entities;
-- +goose StatementEnd
