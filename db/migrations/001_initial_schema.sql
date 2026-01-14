CREATE TABLE IF NOT EXISTS submissions (
    username VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    submission_count INT NOT NULL DEFAULT 1,
    PRIMARY KEY (username, timestamp)
);

CREATE INDEX idx_submissions_timestamp ON submissions(timestamp);
CREATE INDEX idx_submissions_username ON submissions(username);

