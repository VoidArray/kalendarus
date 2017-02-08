.PHONY: run all linux example

run:
	@go run *.go --config-file kalendarus.toml

all:
	@mkdir -p bin/
	@bash --norc -i ./scripts/build.sh

linux:
	@mkdir -p bin/
	@export GOOS=linux && export GOARCH=amd64 && bash --norc -i ./scripts/build.sh

example:
	@php -S 127.0.0.1:8080 -t ./example
