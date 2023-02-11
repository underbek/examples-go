-- +goose Up
-- +goose StatementBegin

-- Rename table
alter table users rename to accounts;

-- +goose StatementEnd

