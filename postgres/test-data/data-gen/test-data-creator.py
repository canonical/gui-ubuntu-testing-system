#!/usr/bin/python3

# This is a throwaway, bad script, used to create some test data.
# It is only being preserved in the repo because it's hard to
# justify not preserving it.

import uuid
import datetime
import hashlib
from random import randrange, choice


utctz = datetime.timezone(datetime.timedelta(hours=0), name="utc")

csv_files = ["jobs.csv", "tests.csv", "users.csv", "reporter.csv"]

jobs_columns = ["uuid", "artifact_url", "tests_repo", "tests_repo_branch", "tests_plans", "image_url", "reporter", "status", "submitted_at", "requester", "debug", "priority"]
tests_columns = ["uuid", "test_case", "vnc_address", "state", "results_url", "updated_at"]
users_columns = ["username", "key", "maximum_priority"]
reporters_columns = ["uuid", "base_reporting_url"]

users = [
    "andersson123",
    "dloose",
    "ashuntu",
    "hk21702",
]

jobs_in_progress = 3
jobs_complete = 10
jobs_pending = 2

jobs_csv = ",".join(jobs_columns) + "\n"
tests_csv = ",".join(tests_columns) + "\n"
users_csv = ",".join(users_columns) + "\n"
reporters_csv = ",".join(reporters_columns) + "\n"


for user in users:
    users_row = [
        user,
        hashlib.sha256(user.encode("utf-8")).hexdigest(),
        str(10),
    ]
    users_csv += ",".join(users_row) + "\n"


default_jobs_data = {
    "uuid": None,
    "artifact_url": "",
    "tests_repo": "https://github.com/canonical/ubuntu-gui-testing.git",
    "tests_repo_branch": "main",
    "tests_plans": None,
    "image_url": "https://cdimage.ubuntu.com/daily-live/current/questing-desktop-amd64.iso",
    "reporter": "test_observer",
    "status": None,
    "submitted_at": None,
    "requester": None,
    "debug": "false",
    "priority": None
}

default_tests_data = {
    "uuid": None,
    "test_case": None,
    "vnc_address": None,
    "state": None,
    "results_url": None,
    "updated_at": None,
}

firefox_data = {
    "artifact_url": "",
    "tests_plans": {
        "tests/firefox-example/plans/extended.yaml": ["Firefox-Example-Basic", "Firefox-Example-New-Tab"],
        "tests/firefox-example/plans/regular.yaml": ["Firefox-Example-Basic"],
    },
    "requester": "andersson123",
}

firmware_updater_data = {
    "artifact_url": "",
    "tests_plans": {
        "tests/firmware-updater/plans/tpm-fde.yaml": ["Firmware-Updater-Tpm-Fde"],
    },
    "requester": "dloose",
}

gnome_shell_data = {
    "artifact_url": "",
    "tests_plans": {
        "tests/gnome-shell/plans/regular.yaml": ["Gnome-Shell-Basic"],
    },
    "requester": "ashuntu",
}

multipass_data = {
    "artifact_url": "",
    "tests_plans": {
        "tests/multipass/plans/regular.yaml": ["Multipass-Basic"],
    },
    "requester": "hk21702",
}

for application in [firefox_data, firmware_updater_data, gnome_shell_data, multipass_data]:
    for jb in range(jobs_in_progress):
        # add the data to the jobs csv
        this_row = []
        for key in jobs_columns:
            if default_jobs_data[key] is None and key in list(application.keys()):
                this_row.append(application[key])
            else:
                this_row.append(default_jobs_data[key])
        job_uuid = str(uuid.uuid4())
        this_row[0] = job_uuid  # uuid
        this_row[4] = '"{' + ','.join(list(application["tests_plans"].keys())) + '}"' # tests_plans
        this_row[7] = "running"  # status
        this_row[8] = str(datetime.datetime.now().astimezone(utctz).isoformat())  # submitted_at
        this_row[11] = str(randrange(10))
        jobs_csv += ",".join(this_row) + "\n"
        # add the data to the tests csv
        for test_plan in application["tests_plans"]:
            for test_case in application["tests_plans"][test_plan]:
                this_test_row = [
                    job_uuid,
                    test_case,
                    f"127.0.0.1:{randrange(5900,6000)}",
                    choice(["requested", "spawning", "spawned", "running"]),
                    "null",
                    str(datetime.datetime.now().astimezone(utctz).isoformat()),
                ]
                tests_csv += ",".join(this_test_row) + "\n"
        # add the data to the reporter csv
        this_reporter_row = [
            job_uuid,
            f"https://tests-api.ubuntu.com/v1/test-executions/{randrange(1000)}",
        ]
        reporters_csv += ",".join(this_reporter_row) + "\n"
    for jb in range(jobs_complete):
        # add complete jobs
        # add the data to the jobs csv
        this_row = []
        for key in jobs_columns:
            if default_jobs_data[key] is None and key in list(application.keys()):
                this_row.append(application[key])
            else:
                this_row.append(default_jobs_data[key])
        job_uuid = str(uuid.uuid4())
        this_row[0] = job_uuid  # uuid
        this_row[4] = '"{' + ','.join(list(application["tests_plans"].keys())) + '}"' # tests_plans
        this_row[7] = choice(["pass", "fail"])
        this_row[8] = str(datetime.datetime.now().astimezone(utctz).isoformat())  # submitted_at
        this_row[11] = str(randrange(10))
        jobs_csv += ",".join(this_row) + "\n"
        # add the data to the tests csv
        for test_plan in application["tests_plans"]:
            for test_case in application["tests_plans"][test_plan]:
                this_test_row = [
                    job_uuid,
                    test_case,
                    f"127.0.0.1:{randrange(5900,6000)}",
                    "fail" if this_row[7] == "fail" else "pass",
                    f"https://guts.ubuntu.com/artifacts/{job_uuid}/",
                    str(datetime.datetime.now().astimezone(utctz).isoformat()),
                ]
                tests_csv += ",".join(this_test_row) + "\n"
        # add the data to the reporter csv
        this_reporter_row = [
            job_uuid,
            f"https://tests-api.ubuntu.com/v1/test-executions/{randrange(1000)}",
        ]
        reporters_csv += ",".join(this_reporter_row) + "\n"
    for jb in range(jobs_pending):
        # add complete jobs
        # add the data to the jobs csv
        this_row = []
        for key in jobs_columns:
            if default_jobs_data[key] is None and key in list(application.keys()):
                this_row.append(application[key])
            else:
                this_row.append(default_jobs_data[key])
        job_uuid = str(uuid.uuid4())
        this_row[0] = job_uuid  # uuid
        this_row[4] = '"{' + ','.join(list(application["tests_plans"].keys())) + '}"' # tests_plans
        this_row[7] = "pending"  # status
        this_row[8] = str(datetime.datetime.now().astimezone(utctz).isoformat())  # submitted_at
        this_row[11] = str(randrange(10))
        jobs_csv += ",".join(this_row) + "\n"

        # add pending jobs
        pass

print("*" * 100)
print("JOBS:")
print("*" * 100)
print(jobs_csv)

print("*" * 100)
print("TESTS:")
print("*" * 100)
print(tests_csv)

print("*" * 100)
print("REPORTERS:")
print("*" * 100)
print(reporters_csv)

print("*" * 100)
print("USERS:")
print("*" * 100)
print(users_csv)

with open("./jobs.csv", "w") as f:
    f.write(jobs_csv)

with open("./tests.csv", "w") as f:
    f.write(tests_csv)

with open("./reporters.csv", "w") as f:
    f.write(reporters_csv)

with open("./users.csv", "w") as f:
    f.write(users_csv)

