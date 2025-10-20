-- create db

CREATE DATABASE guts;

\c guts;

DROP TABLE IF EXISTS currtztable;

SELECT current_setting('TIMEZONE') INTO currtztable;

DO
$do$
BEGIN
    IF (SELECT count(*) = 0 FROM currtztable WHERE current_setting='UTC') THEN
        ALTER DATABASE guts SET timezone TO 'UTC';
    END IF;
END
$do$;

DROP TABLE IF EXISTS currtztable;

-- create tables
CREATE TABLE IF NOT EXISTS jobs (
    uuid VARCHAR(36) PRIMARY KEY NOT NULL,  -- noqa: RF04
    artifact_url VARCHAR(300),
    tests_repo VARCHAR(300) NOT NULL,
    tests_repo_branch VARCHAR(200) NOT NULL,
    tests_plans VARCHAR [],
    image_url VARCHAR(300) NOT NULL,
    reporter VARCHAR(50) NOT NULL,
    status VARCHAR(10) NOT NULL,
    submitted_at TIMESTAMP WITH TIME ZONE NOT NULL,
    requester VARCHAR(50) NOT NULL,
    debug BOOLEAN NOT NULL,
    priority INTEGER NOT NULL,
    CONSTRAINT constrain_status CHECK (status IN (
        'pending',
        'running', 'pass', 'fail'
    ))
);

CREATE TABLE IF NOT EXISTS tests (
    uuid VARCHAR(36) NOT NULL,  -- noqa: RF04
    test_case VARCHAR(100),
    vnc_address VARCHAR(50),
    state VARCHAR(50),
    CONSTRAINT constrain_state CHECK (
        state IN ('requested', 'spawning', 'spawned', 'running', 'pass', 'fail')
    ),
    results_url VARCHAR(300),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT uuid_key FOREIGN KEY (uuid) REFERENCES jobs (uuid)
);

CREATE TABLE IF NOT EXISTS users (
    username VARCHAR(50),
    key VARCHAR(200),  -- stored as sha256 sum  -- noqa: RF04
    maximum_priority INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS reporter (
    uuid VARCHAR(36) NOT NULL,  -- noqa: RF04
    base_reporting_url VARCHAR(300),
    CONSTRAINT uuid_key FOREIGN KEY (uuid) REFERENCES jobs (uuid)
);  -- \n
