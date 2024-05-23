SHELL := /bin/bash
.PHONY: *


help:
	echo "Please execute a specific make command. e.g. make build"


build:
	docker build -t localhost/aws-usage-alerts .

test:
	go test ./...

.SILENT:  # all targets
