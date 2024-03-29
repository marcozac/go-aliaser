#!/bin/sh

set -e

# Generate aliases for uuid package from github.com/google/uuid and write to
# uuid/alias.go file. It also runs go mod tidy to remove unused dependencies.
go run -mod=mod github.com/marcozac/go-aliaser/cmd/aliaser generate \
	--target="uuid" \
	--pattern="github.com/google/uuid" \
	--file="uuid/alias.go" \
	&& go mod tidy
