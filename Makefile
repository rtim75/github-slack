PROJECTNAME=$(shell basename "$(PWD)")

MAKEFLAGS += --silent

all: build deploy

build:
	echo "Building release-check"
	cd release-check && GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -o ../bin/check

deploy:
	echo "Deploying serverless"
	serverless deploy -c serverless.yml -s dev