-- +migrate Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- +migrate StatementBegin
DO $$
BEGIN
    CREATE USER admin WITH PASSWORD 'password' LOGIN;
    EXCEPTION WHEN DUPLICATE_OBJECT THEN RAISE NOTICE 'not creating role transcoder -- it already exists';
END
$$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$
BEGIN
    CREATE USER readonly WITH PASSWORD 'password' LOGIN;
    EXCEPTION WHEN DUPLICATE_OBJECT THEN RAISE NOTICE 'not creating role readonly -- it already exists';
END
$$;
-- +migrate StatementEnd

-- +migrate Down
DROP EXTENSION IF EXISTS "uuid-ossp" CASCADE;
