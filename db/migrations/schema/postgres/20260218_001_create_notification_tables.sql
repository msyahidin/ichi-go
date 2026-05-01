-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS notification_template_overrides (
    id             BIGSERIAL       NOT NULL PRIMARY KEY,
    versions       BIGINT          NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ              DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ              DEFAULT NULL,
    created_by     BIGINT          NOT NULL DEFAULT 0,
    updated_by     BIGINT          NOT NULL DEFAULT 0,
    deleted_by     BIGINT          NOT NULL DEFAULT 0,

    event_slug     VARCHAR(100)    NOT NULL,
    channel        VARCHAR(50)     NOT NULL,
    locale         VARCHAR(10)     NOT NULL DEFAULT 'en',
    title_template VARCHAR(500)             DEFAULT NULL,
    body_template  TEXT                     DEFAULT NULL,
    is_active      BOOLEAN         NOT NULL DEFAULT TRUE
);

CREATE UNIQUE INDEX uniq_active_template_lookup ON notification_template_overrides (event_slug, channel, locale) WHERE deleted_at IS NULL AND is_active = TRUE;

CREATE TABLE IF NOT EXISTS notification_campaigns (
    id              BIGSERIAL       NOT NULL PRIMARY KEY,
    versions        BIGINT          NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ              DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ              DEFAULT NULL,
    created_by      BIGINT          NOT NULL DEFAULT 0,
    updated_by      BIGINT          NOT NULL DEFAULT 0,
    deleted_by      BIGINT          NOT NULL DEFAULT 0,

    delivery_mode   VARCHAR(20)     NOT NULL,
    event_slug      VARCHAR(100)    NOT NULL,
    channels        JSONB           NOT NULL,
    user_target_ids JSONB                    DEFAULT NULL,
    user_exclude_ids JSONB                   DEFAULT NULL,
    locale          VARCHAR(10)     NOT NULL DEFAULT 'en',
    data            JSONB                    DEFAULT NULL,
    meta            JSONB                    DEFAULT NULL,
    scheduled_at    TIMESTAMPTZ              DEFAULT NULL,
    delay_seconds   INT                      DEFAULT NULL,
    status          VARCHAR(20)     NOT NULL DEFAULT 'pending',
    error_message   TEXT                     DEFAULT NULL,
    published_at    TIMESTAMPTZ              DEFAULT NULL
);

CREATE INDEX idx_campaign_status ON notification_campaigns (status, created_at);
CREATE INDEX idx_campaign_event  ON notification_campaigns (event_slug, created_at);

CREATE TABLE IF NOT EXISTS notification_logs (
    id          BIGSERIAL   NOT NULL PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ          DEFAULT NOW(),

    campaign_id BIGINT      NOT NULL,
    user_id     BIGINT      NOT NULL DEFAULT 0,
    channel     VARCHAR(50) NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending',
    error       TEXT                 DEFAULT NULL,
    sent_at     TIMESTAMPTZ          DEFAULT NULL
);

CREATE INDEX idx_log_campaign ON notification_logs (campaign_id);
CREATE INDEX idx_log_user     ON notification_logs (user_id, channel, created_at);
CREATE INDEX idx_log_status   ON notification_logs (status, created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notification_logs;
DROP TABLE IF EXISTS notification_campaigns;
DROP TABLE IF EXISTS notification_template_overrides;
-- +goose StatementEnd
