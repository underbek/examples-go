-- +goose Up
-- +goose StatementBegin

-- UUID extension
create extension if not exists "uuid-ossp";

-- Users table
create table if not exists users
(
    id         uuid                  default uuid_generate_v4(),
    name       varchar(100) not null,
    email      varchar(100) not null,
    phone      varchar(100),
    region     varchar(5)   not null,
    created_at timestamp    not null default now(),
    updated_at timestamp    not null default now(),
    primary key (id),
    unique (email)
);

-- +goose StatementEnd


