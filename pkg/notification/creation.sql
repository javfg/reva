USE cernboxngcopy;

CREATE TABLE `cbox_notifications` (
	`id` INT PRIMARY KEY AUTO_INCREMENT,
	`ref` VARCHAR(3072) UNIQUE NOT NULL,
	`template_name` VARCHAR(320) NOT NULL
);

COMMIT;

CREATE TABLE `cbox_notification_recipients` (
	`id` INT PRIMARY KEY AUTO_INCREMENT,
	`notification_id` INT NOT NULL,
	`user_name` VARCHAR(320) NOT NULL,
	FOREIGN KEY (notification_id)
		REFERENCES cbox_notifications (id)
		ON DELETE CASCADE
);

COMMIT;

CREATE INDEX `cbox_notifications_ix0` ON `cbox_notifications` (`ref`);

CREATE INDEX `cbox_notification_recipients_ix0` ON `cbox_notification_recipients` (`notification_id`);
CREATE INDEX `cbox_notification_recipients_ix1` ON `cbox_notification_recipients` (`user_name`);

-- ocm share

ALTER TABLE cernboxngcopy.oc_share ADD notify_uploads BOOL DEFAULT false;

UPDATE cernboxngcopy.oc_share SET notify_uploads = false;

ALTER TABLE cernboxngcopy.oc_share MODIFY notify_uploads BOOL DEFAULT false NOT NULL;
