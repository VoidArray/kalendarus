.PHONY: run all linux

run:
	@go run *.go --config-file kalendarus.toml

all:
	@mkdir -p bin/
	@bash --norc -i ./scripts/build.sh

linux:
	@mkdir -p bin/
	@export GOOS=linux && export GOARCH=amd64 && bash --norc -i ./scripts/build.sh
