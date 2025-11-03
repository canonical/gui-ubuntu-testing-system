#!/bin/bash

set -ex

# need to delineate between local and hosted deployments
LOCAL=$1
DB_RESET=$2

if [[ "${LOCAL}" != "local" && "${LOCAL}" != "hosted" ]]; then
    printf "%s must be one of ['local', 'hosted']\n" "${LOCAL}"
    exit 1
fi

if [[ "${DB_RESET}" != "yes" && "${DB_RESET}" != "no" ]]; then
    printf "%s must be one of ['yes', 'no']\n" "${DB_RESET}"
    exit 1
fi

if [ -z "${PG_HOST}" ] && [ "${LOCAL}" == "hosted" ]; then printf "\$PG_HOST must be set"; exit 1; fi
if [ -z "${PG_PORT}" ] && [ "${LOCAL}" == "hosted" ]; then printf "\$PG_PORT must be set"; exit 1; fi
if [ -z "${PG_USER}" ] && [ "${LOCAL}" == "hosted" ]; then printf "\$PG_USER must be set"; exit 1; fi

if [[ "${DB_RESET}" == "yes" ]]; then
  if [[ $LOCAL == "local" ]]; then
      sudo -i -u postgres psql -a -c 'DROP DATABASE guts;' || true
  else
      psql --host="${PG_HOST}" --port="${PG_PORT}" --user="${PG_USER}" -a -c 'DROP DATABASE guts;' || true
  fi
fi

CURRDIR=$(pwd)
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PATCH_DIR="${SCRIPT_DIR}/../schema/patches/"
TEST_DATA_DIR="${SCRIPT_DIR}/../test-data/"

PATCHES=$(ls "${PATCH_DIR}")

# TEMP_DIR=$(mktemp -d)
TEMP_DIR="/var/lib/postgresql/data/"
sudo rm -r $TEMP_DIR || true
echo "removed the data directory"
sudo mkdir $TEMP_DIR

sudo sudo cp -br "${PATCH_DIR}/" "${TEMP_DIR}"

sudo chown -R postgres:postgres "${TEMP_DIR}"

for patch in $PATCHES; do
    if [[ $LOCAL == "local" ]]; then
        sudo -i -u postgres psql -a -f "${TEMP_DIR}/patches/${patch}"
    else
        psql --host="${PG_HOST}" --port="${PG_PORT}" --user="${PG_USER}" -a -f "${TEMP_DIR}/patches/${patch}"
    fi
done

# remove everything in the directory but preserve it
sudo rm -r $TEMP_DIR || true
echo "removed the data directory"
sudo mkdir $TEMP_DIR

if [[ "${DB_RESET}" == "yes" ]]; then
  sudo cp -br "${TEST_DATA_DIR}/" "${TEMP_DIR}"
  sudo chown -R postgres:postgres "${TEMP_DIR}"
  if [[ $LOCAL == "local" ]]; then
      sudo -i -u postgres psql -a -f "${TEMP_DIR}/test-data/test-data.sql"
  else
      psql --host="${PG_HOST}" --port="${PG_PORT}" --user="${PG_USER}" -a -f "${TEMP_DIR}/test-data/test-data.sql"
  fi
fi

cd "${CURRDIR}"

sudo rm -r "${TEMP_DIR}" || true
echo "removed the data directory"
