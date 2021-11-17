# Insolvency API

[![GoDoc](https://godoc.org/github.com/companieshouse/insolvency-api?status.svg)](https://godoc.org/github.com/companieshouse/insolvency-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/companieshouse/insolvency-api)](https://goreportcard.com/report/github.com/companieshouse/insolvency-api)

API code and specification for submitting insolvency data to Companies House

## Requirements

In order to run this API locally you will need to install the following:

- [Go](https://golang.org/doc/install)
- [Git](https://git-scm.com/downloads)
- [MongoDB](https://www.mongodb.com/)

## Running Locally with Docker CHS - PLACEHOLDER

Clone Docker CHS Development and follow the steps in the README.

Enable the `insolvency` module

Navigate to http://api.chs.local:<PORT_TO_BE_DECIDED>

Development mode is available for this service in Docker CHS Development.

`./bin/chs-dev development enable insolvency-api`

Swagger documentation is available for this service in the docker CHS development

Navigate to http://api.chs.local/api-docs/chs-delta-api/swagger-ui.html

## Running locally with Docker but without Docker CHS - PLACEHOLDER

Pull image from private CH registry by running docker pull 169942020521.dkr.ecr.eu-west-1.amazonaws.com/local/insolvency-api:latest command or run the following steps to build image locally:

1.  `export SSH_PRIVATE_KEY_PASSPHRASE='[your SSH key passhprase goes here]'` (optional, set only if SSH key is passphrase protected)
2.  `DOCKER_BUILDKIT=0 docker build --build-arg SSH_PRIVATE_KEY="$(cat ~/.ssh/id_rsa)" --build-arg SSH_PRIVATE_KEY_PASSPHRASE -t 169942020521.dkr.ecr.eu-west-1.amazonaws.com/local/insolvency-api .`
3.  `docker run 169942020521.dkr.ecr.eu-west-1.amazonaws.com/local/insolvency-api:latest`

## Running locally without Docker

1. Clone this repository: `go get github.com/companieshouse/insolvency-api`
1. Build the executable: `make build`

## Configuration

| Variable                        | Default | Description             |
| :------------------------------ | :------ | :---------------------- |
| `BIND_ADDR`                     | `-`     | Insolvency API Port     |
| `MONGODB_URL`                   | `-`     | MongoDB URL             |
| `INSOLVENCY_MONGODB_DATABASE`   | `-`     | MongoDB database name   |
| `INSOLVENCY_MONGODB_COLLECTION` | `-`     | MongoDB collection name |

## Swagger UI

Insolvency API specifications schema: [https://github.com/companieshouse/insolvency-api/blob/master/apispec/schema.yml]().

Run the following to view the spec in the browser:

    docker-compose up -d

Spec can be viewed at [http://localhost:8080](http://localhost:8080)
