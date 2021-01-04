PROJECTNAME=$(shell basename "$(PWD)")

MAKEFLAGS += --silent

all: build deploy

build: build-release-check build-release-post

build-release-check:
	echo "Building release-check"
	cd release-check && GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -o ../bin/release-check

build-release-post:
	echo "Building release-post"
	cd release-post && GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -o ../bin/release-post


deploy:
	echo "Deploying serverless"
	serverless deploy -c serverless.yml -s dev