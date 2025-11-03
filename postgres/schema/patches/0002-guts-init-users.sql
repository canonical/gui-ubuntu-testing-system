\c guts

CREATE OR REPLACE FUNCTION add_user_if_not_exists(username NAME, pw TEXT)
RETURNS INTEGER
AS $$
BEGIN
   IF NOT EXISTS ( SELECT FROM pg_roles  
                   WHERE  rolname = username) THEN

        EXECUTE FORMAT('CREATE ROLE "%I" PASSWORD %L', username, pw);
   END IF;
   RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- create users
SELECT add_user_if_not_exists('guts_api', 'guts_api');
SELECT add_user_if_not_exists('guts_spawner', 'guts_spawner');
SELECT add_user_if_not_exists('guts_scheduler', 'guts_scheduler');
SELECT add_user_if_not_exists('guts_runner', 'guts_runner');
SELECT add_user_if_not_exists('guts_reporter', 'guts_reporter');

-- create permissions

-- api permissions
GRANT SELECT, INSERT ON jobs TO guts_api;
GRANT SELECT, INSERT ON tests TO guts_api;
GRANT SELECT ON users TO guts_api;
ALTER USER guts_api WITH LOGIN;

-- spawner permissions
GRANT SELECT, UPDATE ON tests TO guts_spawner;
GRANT SELECT ON jobs TO guts_spawner;
ALTER USER guts_spawner WITH LOGIN;

-- scheduler permissions
GRANT SELECT, UPDATE, DELETE ON jobs TO guts_scheduler;
GRANT SELECT, INSERT, UPDATE, DELETE ON tests TO guts_scheduler;
GRANT SELECT, DELETE ON reporter TO guts_scheduler;
ALTER USER guts_scheduler WITH LOGIN;

-- runner permissions
GRANT SELECT, UPDATE ON tests TO guts_runner;
GRANT SELECT ON jobs TO guts_runner;
ALTER USER guts_runner WITH LOGIN;

-- reporter permissions
GRANT SELECT, INSERT, UPDATE ON reporter TO guts_reporter;
ALTER USER guts_reporter WITH LOGIN; -- \n
