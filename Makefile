## Before we start test that we have the mandatory executables available
EXECS = go docker
K := $(foreach exec,$(EXECS),\
$(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH, consider installing $(exec)")))
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
export CGO_ENABLED=0

.PHONY: clean

.ONESHELL:

compile:
	go build weather.go

build: compile
	docker build -t weather .


clean:
	docker rmi -f weather
	rm -vf weather