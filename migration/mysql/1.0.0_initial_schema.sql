-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS schedules (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    date         DATE NOT NULL,        -- date of termination schedule, in local time zone
    time         DATETIME NOT NULL,    -- time in UTC. Because of time difference, may differ from date
    team          VARCHAR(512) NOT NULL,
    account      VARCHAR(100) NOT NULL,
    region       VARCHAR(50)  NOT NULL, -- use blank string to indicate not present
    stack        VARCHAR(255) NOT NULL, -- use blank string to indicate not present
    team      VARCHAR(768) NOT NULL, -- use blank string to indicate not present
    INDEX date_index (date)
    )
ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS terminations (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    team          VARCHAR(512) NOT NULL,
    account      VARCHAR(100) NOT NULL,
    stack        VARCHAR(255) NOT NULL,
    team      VARCHAR(768) NOT NULL,
    region       VARCHAR(50) NOT NULL,
    asg          VARCHAR(1000) NOT NULL,
    employee_id  VARCHAR(48) NOT NULL,
    fired_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    leashed      BOOLEAN NOT NULL DEFAULT FALSE,
    INDEX app_fired_at_index (app,fired_at)
    )
ENGINE=InnoDB;


-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE schedules;
DROP TABLE terminations;
