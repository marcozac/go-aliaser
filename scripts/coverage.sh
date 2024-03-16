#!/usr/bin/env bash

cd "$(dirname "$0")/.." \
	|| (echo "failed to cd to project root" && exit 1)

go test -race \
	-covermode=atomic \
	-coverpkg=./... \
	-coverprofile=cover.out \
	./... \
	&& go tool cover -html=cover.out -o=cover.html
