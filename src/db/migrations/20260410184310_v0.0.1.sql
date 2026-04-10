-- Create "tasks" table
CREATE TABLE `tasks` (
  `id` text NOT NULL,
  `active` integer NOT NULL DEFAULT 1,
  `title` text NOT NULL,
  `star` integer NOT NULL DEFAULT 0,
  `rrule` text NULL,
  `dtstart` datetime NULL,
  `until` datetime NULL,
  `count` integer NULL,
  `next_occurrence` datetime NULL,
  `created_at` datetime NULL DEFAULT (CURRENT_TIMESTAMP),
  `updated_at` datetime NULL DEFAULT (CURRENT_TIMESTAMP),
  `completed_at` datetime NULL,
  PRIMARY KEY (`id`),
  CHECK (active IN (0, 1)),
  CHECK (star in (0, 1))
);
-- Create index "idx_task_next_occurrence" to table: "tasks"
CREATE INDEX `idx_task_next_occurrence` ON `tasks` (`next_occurrence`, `active`);
-- Create "schema_migrations" table
CREATE TABLE `schema_migrations` (
  `version` text NULL,
  `applied_at` datetime NULL DEFAULT (CURRENT_TIMESTAMP),
  PRIMARY KEY (`version`)
);
