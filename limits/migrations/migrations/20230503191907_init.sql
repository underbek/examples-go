-- +goose Up
-- +goose StatementBegin
CREATE TYPE limit_type AS ENUM ('min_amount','max_amount','total_amount','total_count');
CREATE TYPE period_type AS ENUM ('calendar_day','calendar_week','calendar_month');
CREATE TYPE operation_status AS ENUM ('new','pending','committed','rollback');
CREATE TYPE action_type AS ENUM ('increase','decrease');

CREATE TABLE IF NOT EXISTS limits
(
    id         bigserial primary key,
    hash       varchar                 not null,
    currency   varchar                 not null,
    meta       jsonb                   not null,
    limit_type limit_type              not null,
    value      varchar                 not null,
    period     period_type,
    timezone   varchar,
    created_at timestamp default now() not null,
    updated_at timestamp default now() not null,
    deleted_at timestamp
);

CREATE TABLE IF NOT EXISTS counters
(
    id         bigserial primary key,
    hash       varchar                 not null,
    limit_id   bigint                  not null references limits (id) on delete cascade,
    value      varchar                 not null,
    start_time timestamp               not null,
    end_time   timestamp               not null,
    created_at timestamp default now() not null,
    updated_at timestamp default now() not null,
    deleted_at timestamp
);

CREATE TABLE IF NOT EXISTS context
(
    id         bigserial primary key,
    meta       jsonb                   not null,
    created_at timestamp default now() not null,
    updated_at timestamp default now() not null
);

CREATE TABLE IF NOT EXISTS operations
(
    id         bigserial primary key,
    context_id bigint                  not null references context (id) on delete cascade,
    value      varchar                 not null,
    currency   varchar                 not null,
    status     operation_status        not null,
    created_at timestamp default now() not null,
    updated_at timestamp default now() not null
);

CREATE TABLE IF NOT EXISTS operation_to_counter
(
    counter_id   bigint                         not null references counters (id) on delete cascade,
    operation_id bigint                         not null references operations (id) on delete cascade,
    action_type  action_type default 'increase' not null,
    primary key (counter_id, operation_id)
);

CREATE INDEX IF NOT EXISTS idx_limits_hash ON limits USING hash (hash);
CREATE INDEX IF NOT EXISTS idx_limits_meta ON limits USING GIN (meta jsonb_path_ops);

CREATE INDEX IF NOT EXISTS idx_counters_hash ON counters USING hash (hash);
CREATE INDEX IF NOT EXISTS idx_counters_limit_id ON counters USING hash (limit_id);

CREATE INDEX IF NOT EXISTS idx_context_meta ON context USING GIN (meta jsonb_path_ops);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_context_meta;
DROP INDEX IF EXISTS idx_counters_limit_id;
DROP INDEX IF EXISTS idx_counters_hash;
DROP INDEX IF EXISTS idx_limits_meta;
DROP INDEX IF EXISTS idx_limits_hash;

DROP TABLE IF EXISTS operation_to_counter;
DROP TABLE IF EXISTS operations;
DROP TABLE IF EXISTS context;
DROP TABLE IF EXISTS counters;
DROP TABLE IF EXISTS limits;

DROP TYPE IF EXISTS limit_type;
DROP TYPE IF EXISTS period_type;
DROP TYPE IF EXISTS operation_status;
DROP TYPE IF EXISTS action_type;
-- +goose StatementEnd
