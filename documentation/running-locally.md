# Running Locally

Running guts locally is nice and simple. Make sure that you have postgres installed and running, and note the port number in the systemd service for postgres. If postgres isn't running on `5433`, you'll need to change the config files for these applications.

## Bootstrap the DB

First, you need to bootstrap the DB with the appropriate schema. From the root of GUTS, run:

```

./postgres/scripts/bootstrap-db.sh local no

```

## Adding your user authentication

Run the following commands:


```

# bash
sudo su postgres
psql -d guts
# next command in psql interactive shell
INSERT INTO users (username, key, maximum_priority) VALUES ('username', 'fa1380650c2239c150f231e3ef37627cbf5ad531782ea2adce2b57478d609f95', 10)

```

## Running applications

### API

Run the following commands:

```

cd guts/cmd/api/
go build -o api && ./api -cfg-path ../../guts-api.yaml

```

### Scheduler

Run the following commands:

```

cd guts/cmd/scheduler/
go build -o scheduler && ./scheduler -cfg-path ../../scheduler/guts-scheduler-local.yaml

```

### Spawner

`qemu` needs to be installed and in `$PATH` on the machine running the Spawner.

Run the following commands:

```

cd guts/cmd/scheduler/
go build -o spawner && ./spawner -cfg-path ../../spawner/guts-spawner.yaml -host localhost -port 5900

```

### Runner

`yarf` needs to be installed and in `$PATH` on the machine running the Runner. Yarf installation instructions are [here](https://github.com/canonical/yarf/?tab=readme-ov-file#installation).

Run the following commands:

```

cd guts/cmd/scheduler/
go build -o runner && ./runner -cfg-path ../../runner/guts-runner-local.yaml

```

You'll see, when the runner completes, a storage url is printed in the runner logs when yarf exits and the artifacts are uploaded to the storage. To make that storage url work, you need to (in a new terminal):

```

cd /srv/data/
php -S localhost:9999 .

```

This will serve files from that directory, which is used by the api when collating test results. You can also download the tarballs yourself if you're interested.

## Submitting test requests

Oop! Everything is running. Awesome. Now, you can submit a valid test request (i.e. the test should actually run and succeed) with the following python script:

```

#!/usr/bin/python3

import requests

POST_JSON = {"artifact_url": None, "tests_repo": "https://github.com/canonical/ubuntu-gui-testing.git", "tests_repo_branch": "main", "tests_plans": ["tests/desktop-installer/plans/noble-smoke.yaml"], "testbed": "https://releases.ubuntu.com/noble/ubuntu-24.04.3-desktop-amd64.iso", "debug": False, "priority": 1, "reporter": ""}

POST_HEADERS = {
    "Content-Type": "application/json",
    "X-Api-Key": "4c126f75-c7d8-4a89-9370-f065e7ff4208",
}

BASE_API_URL = "http://localhost:8080/"
REQ_URL = f"{BASE_API_URL}request/"

r = requests.post(
    REQ_URL,
    json=POST_JSON,
    headers=POST_HEADERS,
)
print(r)
print(r.content.decode("utf-8"))

```

## Viewing the test

Install the `xtightvncviewer` package. Then, view your running test with:

```

vncviewer $host:5900 -viewonly

```

## Conclusion

And that should be it!
