# gui-ubuntu-testing-system - GUTS

GUTS is a job scheduler and runner for [yarf](https://github.com/canonical/yarf) based automated GUI testing. You request a test via a web api, and it runs on the testbed you specify, against artifacts you specify. It is still in development.

Infrastructure to run tests defined in [ubuntu-gui-testing](https://github.com/canonical/ubuntu-gui-testing).

You can see the architecture diagram [here](documentation/architecture/README.md).

## Development

To run unit tests, run: 

```
./scripts/run-tests.bash
```

The go module is in the `guts/` directory. Each sub-directory is a package. Each sub-directory corresponds to an application or a group of utilities.

In each `guts/$application/` directory, there will be a dummy config file that also suffices for running unit tests locally.

You can build the corresponding executables in the `guts/cmd/$application/` directories.

The `postgres/` directory contains all database related things; schemas, test data, helper scripts, etc.

## Running Locally

Build all of the executables by going to each `guts/cmd/$application/` directory and running `go build -o $application`.

Make sure you have postgres installed locally, and run:

```
./postgres/scripts/bootstrap-db.sh local no
```

If you're running postgres as part of a charmed juju environment, the command would be:

```
./postgres/scripts/bootstrap-db.sh hosted no
```

But you must set the following environment variables first:

```
PG_HOST
PG_PORT
PG_USER
```

Then, start all the guts applications (the reporter is optional), and you will find the api running at:

```
localhost:8080
```

## Applications

### API

The api is pretty self explanatory, to examine the endpoints, see the OpenAPI [spec](readme/openapi-spec).

Via the API you requests tests and monitor their results.

### Scheduler

The scheduler is an application which:
- Handles new job requests by writing them to the tests table
- Updates the complete jobs when all the individual tests have finished
- Resets the state for tests that have a failing runner or spawner
  process
- Removes objects in the object storage older than a specified duration

### Spawner

The spawner waits for tests that need a testbed, and then spawns a testbed for said test.

### Runner

The runner application runs tests with `yarf` on testbeds provided by the `spawner` application, as specified by the job request sent to the api.
