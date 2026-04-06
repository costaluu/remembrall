-- Create "tasks" table
CREATE TABLE `tasks` (
  `id` text NOT NULL,
  `active` integer NOT NULL DEFAULT 1,
  `title` text NOT NULL,
  `dtstart` datetime NULL,
  `rrule` text NULL,
  `until` datetime NULL,
  `count` integer NULL,
  `next_occurrence` datetime NULL,
  `created_at` datetime NULL DEFAULT (CURRENT_TIMESTAMP),
  `updated_at` datetime NULL DEFAULT (CURRENT_TIMESTAMP),
  PRIMARY KEY (`id`),
  CHECK (active IN (0, 1))
);
-- Create index "idx_task_next_occurrence" to table: "tasks"
CREATE INDEX `idx_task_next_occurrence` ON `tasks` (`next_occurrence`, `active`);
-- Create "completions" table
CREATE TABLE `completions` (
  `id` text NOT NULL,
  `task_id` text NOT NULL,
  `occurrence_date` datetime NOT NULL,
  `completed_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`task_id`) REFERENCES `tasks` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_completion_task" to table: "completions"
CREATE INDEX `idx_completion_task` ON `completions` (`task_id`, `occurrence_date`);
-- Create "schema_migrations" table
CREATE TABLE `schema_migrations` (
  `version` text NULL,
  `applied_at` datetime NULL DEFAULT (CURRENT_TIMESTAMP),
  PRIMARY KEY (`version`)
);
