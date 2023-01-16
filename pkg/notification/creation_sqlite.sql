CREATE TABLE `cbox_notifications` (
        `id` INTEGER PRIMARY KEY AUTOINCREMENT,
        `ref` VARCHAR(3072) UNIQUE NOT NULL,
        `template_name` VARCHAR(320) NOT NULL
);

COMMIT;

CREATE TABLE `cbox_notification_recipients` (
	`id` INTEGER PRIMARY KEY AUTOINCREMENT,
	`notification_id` INTEGER NOT NULL,
	`user_name` VARCHAR(320) NOT NULL,
	FOREIGN KEY (notification_id)
		REFERENCES cbox_notifications (id)
		ON DELETE CASCADE
);

COMMIT;

CREATE INDEX `cbox_notifications_ix0` ON `cbox_notifications` (`ref`);

CREATE INDEX `cbox_notification_recipients_ix0` ON `cbox_notification_recipients` (`notification_id`);
CREATE INDEX `cbox_notification_recipients_ix1` ON `cbox_notification_recipients` (`user_name`);

COMMIT;

INSERT INTO `cbox_notifications` (`id`, `ref`, `template_name`) VALUES (1, "notification-test", "notification-template-test")
INSERT INTO `cbox_notification_recipients` (`id`, `notification_id`, `user_name`) VALUES (1, 1, "jdoe"), (2, 1, "testuser")

COMMIT;
