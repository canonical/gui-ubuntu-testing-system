\c guts;

COPY jobs (
    uuid,
    artifact_url,
    tests_repo,
    tests_repo_branch,
    tests_plans,
    image_url,
    reporter,
    status,
    submitted_at,
    requester,
    debug,
    priority
) FROM '/var/lib/postgresql/data/test-data/jobs.csv' DELIMITER ',' CSV HEADER;

COPY tests (
    uuid,
    test_case,
    vnc_address,
    state,
    results_url,
    updated_at,
    plan,
    commit_hash
) FROM '/var/lib/postgresql/data/test-data/tests.csv' DELIMITER ',' CSV HEADER;

COPY users (
    username, key, maximum_priority
) FROM '/var/lib/postgresql/data/test-data/users.csv' DELIMITER ',' CSV HEADER;

COPY reporter (
    uuid, base_reporting_url
) FROM '/var/lib/postgresql/data/test-data/reporters.csv' DELIMITER ',' CSV HEADER; -- noqa:disable=layout.long_lines
