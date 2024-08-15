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

test-cov:
	go test -v -coverprofile tmp_cov.out ./...
	go tool cover -html tmp_cov.out -o tmp_cov.html
	open tmp_cov.html

dev: dev-config-exists
	cd deploy/dev && \
	terraform init && \
	terraform apply --auto-approve

dev-config-exists:
	filepath="deploy/dev/config.yaml" && \
	if [ ! -f $$filepath ]; then \
        echo "$$filepath did not exist. Please create it" && exit 1; \
    fi

dev-destroy:
	cd deploy/dev && \
	terraform apply --auto-approve -destroy

.SILENT:  # all targets
