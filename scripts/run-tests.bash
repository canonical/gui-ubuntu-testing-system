#!/bin/bash
set -e

export GIN_MODE=release

runtests(){
  go test ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...
  go-test-coverage --config=$1
}

cleanup(){
  unset GIN_MODE
  unset GUTS_CFG_PATH
}

trap cleanup EXIT

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
BASE_DIR="${SCRIPT_DIR}/../"
GUTS_DIR="${BASE_DIR}guts"

#################################
# Test packages
git lfs pull
export GUTS_CFG_PATH=../guts-api.yaml

#################################
# Test the API functionality
cd "${GUTS_DIR}/api/"
runtests ./../.testcoverage.yml

# Test the database functionality
cd "${GUTS_DIR}/database/"
runtests ./../.testcoverage.yml

# Test the spawner functionality
cd "${GUTS_DIR}/spawner/"
runtests ./../.testcoverage.yml

# Test the utils functionality
cd "${GUTS_DIR}/utils/"
runtests ./../.testcoverage.yml

#################################
# Test executables
export GUTS_CFG_PATH=../../guts-api.yaml

# Test the api executable
cd "${GUTS_DIR}/cmd/api/"
runtests ./../../.testcoverage.yml

# Test the spawner executable
cd "${GUTS_DIR}/cmd/spawner/"
runtests ./../../.testcoverage.yml
