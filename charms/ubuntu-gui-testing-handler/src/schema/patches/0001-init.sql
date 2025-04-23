CREATE SCHEMA IF NOT EXISTS test_requests;
CREATE SCHEMA IF NOT EXISTS spawn_requests;
CREATE SCHEMA IF NOT EXISTS results;

-- test_requests
CREATE TABLE test_requests.test_requests (
    uuid VARCHAR[36] PRIMARY KEY NOT NULL,  -- uuid4
    family VARCHAR[12],
    ADD CONSTRAINT chk_family CHECK (family in ('snap','deb','image')),
    name VARCHAR[50],
    version VARCHAR[50],
    test_plan VARCHAR[50],
    os VARCHAR[20],
    ADD CONSTRAINT chk_os CHECK (os in ('ubuntu','ubuntu-core','ubuntu-server')),
    series VARCHAR[20],
    execution_stage VARCHAR[20],
    image_url VARCHAR[200],
    test_observer_test_id INTEGER
);

-- spawn_requests
CREATE TABLE spawn_requests.spawn_requests (
    id SERIAL PRIMARY KEY,
    uuid VARCHAR[36] REFERENCES test_requests.test_requests(uuid) NOT NULL,
    tpm BOOLEAN NOT NULL,
    image_type VARCHAR[30],
    ADD CONSTRAINT chk_image_type CHECK (image_type in ('live','pre-installed')),
    test_case VARCHAR[120],
    state VARCHAR[7],
    ADD CONSTRAINT chk_state CHECK (state in ('new','waiting','running', 'killme')),
    vnc_host VARCHAR[45],
    vnc_port INTEGER,
    ADD CONSTRAINT chk_vnc_port_range CHECK (vnc_port BETWEEN 5900 AND 5999)
);

-- results
CREATE TABLE results.results (
    uuid VARCHAR[36] REFERENCES test_requests.test_requests(uuid) PRIMARY KEY NOT NULL,
    artifacts_url VARCHAR[200],
    status VARCHAR[10],
    ADD CONSTRAINT chk_status CHECK (status in ('PASS','FAIL')),
    individual_results NVARCHAR(MAX),  -- {'test-case-1': 'pass', 'test-case-2': 'fail', 'etc': 'fail'}
    ADD CONSTRAINT check_json CHECK ( ISJSON(individual_results)>0 )
);


