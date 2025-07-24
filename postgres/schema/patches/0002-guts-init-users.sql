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
GRANT INSERT ON jobs TO guts_api;

-- spawner permissions
GRANT UPDATE ON tests TO guts_spawner;

-- scheduler permissions
GRANT UPDATE ON jobs TO guts_scheduler;
GRANT INSERT, UPDATE ON tests TO guts_scheduler;

-- runner permissions
GRANT UPDATE ON tests TO guts_runner;

-- reporter permissions
GRANT INSERT, UPDATE ON reporter TO guts_reporter; -- \n
