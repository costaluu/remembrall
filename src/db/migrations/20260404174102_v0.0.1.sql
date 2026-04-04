-- Create "reminders" table
CREATE TABLE `reminders` (
  `id` text NOT NULL,
  `title` text NOT NULL,
  `dtstart` datetime NOT NULL,
  `rrule` text NULL,
  `until` datetime NULL,
  `count` integer NULL,
  `next_trigger` datetime NOT NULL,
  `created_at` datetime NULL DEFAULT (CURRENT_TIMESTAMP),
  `updated_at` datetime NULL DEFAULT (CURRENT_TIMESTAMP),
  `active` boolean NULL DEFAULT 1,
  PRIMARY KEY (`id`),
  CHECK (active IN (0, 1))
);
-- Create index "idx_reminder_next_trigger" to table: "reminders"
CREATE INDEX `idx_reminder_next_trigger` ON `reminders` (`next_trigger`, `active`);
-- Create index "idx_reminder_dtstart" to table: "reminders"
CREATE INDEX `idx_reminder_dtstart` ON `reminders` (`dtstart`);
-- Create "schema_migrations" table
CREATE TABLE `schema_migrations` (
  `version` text NULL,
  `applied_at` datetime NULL DEFAULT (CURRENT_TIMESTAMP),
  PRIMARY KEY (`version`)
);
-- Create "exceptions" table
CREATE TABLE `exceptions` (
  `id` text NOT NULL,
  `reminder_id` text NOT NULL,
  `original_date` datetime NOT NULL,
  `type` text NOT NULL,
  `new_date` datetime NULL,
  `new_title` text NULL,
  `created_at` datetime NULL DEFAULT (CURRENT_TIMESTAMP),
  `updated_at` datetime NULL DEFAULT (CURRENT_TIMESTAMP),
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`reminder_id`) REFERENCES `reminders` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (type IN ('cancelled', 'modified'))
);
-- Create index "idx_exception_unique" to table: "exceptions"
CREATE UNIQUE INDEX `idx_exception_unique` ON `exceptions` (`reminder_id`, `original_date`);
-- Create index "idx_exception_reminder" to table: "exceptions"
CREATE INDEX `idx_exception_reminder` ON `exceptions` (`reminder_id`);
