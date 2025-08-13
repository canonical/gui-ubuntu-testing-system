#!/bin/bash

set -ex

# need to delineate between local and hosted deployments
LOCAL=$1

if [[ "${LOCAL}" != "local" && "${LOCAL}" != "hosted" ]]; then
    printf "%s must be one of ['local', 'hosted']\n" "${LOCAL}"
    exit 1
fi


if [ -z "${PG_HOST}" ] && [ "${LOCAL}" == "hosted" ]; then printf "\$PG_HOST must be set"; exit 1; fi
if [ -z "${PG_PORT}" ] && [ "${LOCAL}" == "hosted" ]; then printf "\$PG_PORT must be set"; exit 1; fi
if [ -z "${PG_USER}" ] && [ "${LOCAL}" == "hosted" ]; then printf "\$PG_USER must be set"; exit 1; fi

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PATCH_DIR="${SCRIPT_DIR}/../schema/patches/"

PATCHES=$(ls "${PATCH_DIR}")

TEMP_DIR=$(mktemp -d)

cp -br "${PATCH_DIR}/" "${TEMP_DIR}"

sudo chown -R postgres:postgres "${TEMP_DIR}"

for patch in $PATCHES; do
    if [[ $LOCAL == "local" ]]; then
        sudo -i -u postgres psql -a -f "${TEMP_DIR}/patches/${patch}"
    else
        psql --host="${PG_HOST}" --port="${PG_PORT}" --user="${PG_USER}" -a -f "${TEMP_DIR}/patches/${patch}"
    fi
done

sudo rm -r "${TEMP_DIR}"

