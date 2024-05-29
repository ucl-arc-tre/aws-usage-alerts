SHELL := /bin/bash
.PHONY: *


help:
	echo "Please execute a specific make command. e.g. make build"


build:
	docker build -t localhost/aws-usage-alerts .

test:
	go test ./...

dev:
	cd deploy/dev && \
	terraform init && \
	terraform apply --auto-approve

dev-destroy:
	cd deploy/dev && \
	terraform apply --auto-approve -destroy

.SILENT:  # all targets
