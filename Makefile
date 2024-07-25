SHELL := /bin/bash
.PHONY: *


help:
	echo "Please execute a specific make command. e.g. make build"


build:
	docker build -t localhost/aws-usage-alerts .

test: test-unit test-integration

test-unit:
	go test ./...

test-integration:
	cd test && \
	docker compose build aws-usage-alerts && \
	docker compose run test && \
	docker compose down

dev:
	cd deploy/dev && \
	terraform init && \
	terraform apply --auto-approve

dev-destroy:
	cd deploy/dev && \
	terraform apply --auto-approve -destroy

.SILENT:  # all targets
