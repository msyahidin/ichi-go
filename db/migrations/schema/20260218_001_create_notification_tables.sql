-- +goose Up
-- +goose StatementBegin

-- notification_template_overrides
-- Optional DB overrides for notification copy per (event_slug, channel, locale).
-- The Go template registry (pkg/notification/template) is the source of truth.
-- Insert rows here only when you want to update copy without a redeploy.
CREATE TABLE IF NOT EXISTS `notification_template_overrides` (
    `id`             BIGINT          NOT NULL AUTO_INCREMENT,
    `versions`       BIGINT          NOT NULL DEFAULT 0,
    `created_at`     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     DATETIME                 DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`     DATETIME                 DEFAULT NULL,
    `created_by`     BIGINT          NOT NULL DEFAULT 0,
    `updated_by`     BIGINT          NOT NULL DEFAULT 0,
    `deleted_by`     BIGINT          NOT NULL DEFAULT 0,

    `event_slug`     VARCHAR(100)    NOT NULL COMMENT 'Must match a slug registered in the Go TemplateRegistry',
    `channel`        VARCHAR(50)     NOT NULL COMMENT 'email | push | sms | in_app | webhook',
    `locale`         VARCHAR(10)     NOT NULL DEFAULT 'en',
    `title_template` VARCHAR(500)             DEFAULT NULL COMMENT 'Go text/template string — overrides Go default',
    `body_template`  TEXT                     DEFAULT NULL COMMENT 'Go text/template string — overrides Go default',
    `is_active`      TINYINT(1)      NOT NULL DEFAULT 1,

    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_template_override` (`event_slug`, `channel`, `locale`, `deleted_at`),
    INDEX `idx_template_lookup` (`event_slug`, `channel`, `locale`, `is_active`, `deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='Optional runtime copy overrides for notification templates';


-- notification_campaigns
-- One record per POST /api/notifications/send call.
-- Tracks lifecycle: pending → published | failed.
CREATE TABLE IF NOT EXISTS `notification_campaigns` (
    `id`              BIGINT          NOT NULL AUTO_INCREMENT,
    `versions`        BIGINT          NOT NULL DEFAULT 0,
    `created_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`      DATETIME                 DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`      DATETIME                 DEFAULT NULL,
    `created_by`      BIGINT          NOT NULL DEFAULT 0,
    `updated_by`      BIGINT          NOT NULL DEFAULT 0,
    `deleted_by`      BIGINT          NOT NULL DEFAULT 0,

    `delivery_mode`   VARCHAR(20)     NOT NULL COMMENT 'blast | user',
    `event_slug`      VARCHAR(100)    NOT NULL COMMENT 'Matches a registered Go EventTemplate slug',
    `channels`        JSON            NOT NULL COMMENT '["email","push"]',
    `user_target_ids` JSON                     DEFAULT NULL COMMENT 'Target user IDs — required for delivery_mode=user',
    `user_exclude_ids` JSON                    DEFAULT NULL COMMENT 'User IDs always excluded from delivery',
    `locale`          VARCHAR(10)     NOT NULL DEFAULT 'en',
    `data`            JSON                     DEFAULT NULL COMMENT 'Template variables passed to text/template',
    `meta`            JSON                     DEFAULT NULL COMMENT 'Operational metadata (trace_id, source, etc.)',
    `scheduled_at`    DATETIME                 DEFAULT NULL COMMENT 'Absolute scheduled delivery time (UTC)',
    `delay_seconds`   INT UNSIGNED             DEFAULT NULL COMMENT 'Relative delay in seconds (max 2147483)',
    `status`          VARCHAR(20)     NOT NULL DEFAULT 'pending' COMMENT 'pending | processing | published | failed',
    `error_message`   TEXT                     DEFAULT NULL,
    `published_at`    DATETIME                 DEFAULT NULL COMMENT 'When the message was successfully queued',

    PRIMARY KEY (`id`),
    INDEX `idx_campaign_status` (`status`, `created_at`),
    INDEX `idx_campaign_event`  (`event_slug`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='One record per notification send API call — campaign lifecycle tracking';


-- notification_logs
-- Per-user, per-channel delivery attempt record.
-- Written by consumers after each dispatch() call.
CREATE TABLE IF NOT EXISTS `notification_logs` (
    `id`          BIGINT      NOT NULL AUTO_INCREMENT,
    `created_at`  DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  DATETIME             DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    `campaign_id` BIGINT      NOT NULL COMMENT 'FK → notification_campaigns.id',
    `user_id`     BIGINT      NOT NULL DEFAULT 0 COMMENT '0 for blast notifications',
    `channel`     VARCHAR(50) NOT NULL COMMENT 'email | push | sms | in_app | webhook',
    `status`      VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT 'pending | sent | failed | skipped',
    `error`       TEXT                 DEFAULT NULL COMMENT 'Error message if status=failed',
    `sent_at`     DATETIME             DEFAULT NULL,

    PRIMARY KEY (`id`),
    INDEX `idx_log_campaign` (`campaign_id`),
    INDEX `idx_log_user`     (`user_id`, `channel`, `created_at`),
    INDEX `idx_log_status`   (`status`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='Per-user, per-channel delivery attempt records';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `notification_logs`;
DROP TABLE IF EXISTS `notification_campaigns`;
DROP TABLE IF EXISTS `notification_template_overrides`;
-- +goose StatementEnd
