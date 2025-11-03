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
DB_BOOTSTRAP_SCRIPT="${BASE_DIR}/postgres/scripts/bootstrap-db.sh"

echo "bootstrapping DB!"
$DB_BOOTSTRAP_SCRIPT local yes

git lfs pull

#################################
# package tests
cd "${GUTS_DIR}/"
export GUTS_CFG_PATH=../guts-api.yaml
pkgs=$(ls -d */ | grep -v cmd)
for pkg in $pkgs; do
  cd "${GUTS_DIR}/${pkg}"
  echo "****************************************************"
  echo "running ${pkg} tests and go fmt"
  go fmt
  runtests ./../.testcoverage.yml
done

#################################
# Test executables
cd "${GUTS_DIR}/cmd"
export GUTS_CFG_PATH=../../guts-api.yaml
excs=$(ls -d */)
for dr in $excs; do
  cd "${GUTS_DIR}/cmd/${dr}"
  echo "****************************************************"
  echo "running cmd/${dr} tests and go fmt"
  go fmt
  runtests ./../../.testcoverage.yml
done
