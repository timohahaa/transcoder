-- +migrate Up
CREATE SCHEMA IF NOT EXISTS transcoder;

CREATE TABLE IF NOT EXISTS transcoder.queue (
      task_id         UUID      NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY
    , source          JSONB     NOT NULL DEFAULT '{}'::jsonb
    , status          TEXT      NOT NULL DEFAULT 'pending'
    , encoder         TEXT      NOT NULL DEFAULT 'auto'
    , routing         TEXT      NOT NULL DEFAULT ''
    , hostname        TEXT      NOT NULL DEFAULT ''
    , duration        REAL      NOT NULL
    , file_size       BIGINT    NOT NULL
    , settings        JSONB     NOT NULL DEFAULT '{}'
    , error           TEXT      NOT NULL DEFAULT ''
    , created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    , updated_at      TIMESTAMP
    , deleted_at      TIMESTAMP

    , CONSTRAINT transcoder_queue_check_status CHECK ( status IN (
            'pending'
            , 'waiting-splitting'
            , 'splitting'
            , 'encoding'
            , 'waiting-assembling'
            , 'assembling'
            , 'canceled'
            , 'error'
            , 'done'
        )
    )
    , CONSTRAINT transcoder_queue_check_encoder CHECK ( encoder IN (
            'auto', 'cpu', 'gpu'
        )
    )
);

GRANT USAGE                  ON               SCHEMA transcoder TO admin, readonly;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA transcoder TO admin;

-- +migrate Down
DROP SCHEMA IF EXISTS transcoder CASCADE;
