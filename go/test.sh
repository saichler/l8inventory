#!/usr/bin/env bash

# Fail on errors and don't open cover file
set -e
# clean up
rm -rf go.mod
rm -rf go.sum
rm -rf vendor

# fetch dependencies
go mod init
GOPROXY=direct GOPRIVATE=github.com go mod tidy
go mod vendor
cp ./vendor/github.com/saichler/l8utils/go/utils/resources/build-test-security.sh .
chmod +x ./build-test-security.sh
rm -rf vendor
./build-test-security.sh
rm -rf ./build-test-security.sh
go mod vendor

# Run unit tests with coverage
go test -tags=unit -v -coverpkg=./inv/... -coverprofile=cover.html ./... --failfast

#rm -rf ./tests/loader.so

# Open the coverage report in a browser
go tool cover -html=cover.html
