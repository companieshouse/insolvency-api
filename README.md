# Insolvency API

[![GoDoc](https://godoc.org/github.com/companieshouse/insolvency-api?status.svg)](https://godoc.org/github.com/companieshouse/insolvency-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/companieshouse/insolvency-api)](https://goreportcard.com/report/github.com/companieshouse/insolvency-api)

API code and specification for submitting insolvency data to Companies House

## Requirements

The recommended way to run this API is via the CHS Docker development mechanism (see [here](https://github.com/companieshouse/docker-chs-development) for more details).

In order to run this API locally, however, you would need to install the following:

- [Go](https://golang.org/doc/install)
- [Git](https://git-scm.com/downloads)
- [MongoDB](https://www.mongodb.com/)

## Running Locally with Docker CHS

Clone Docker CHS Development and follow the steps in the README.

Enable the `insolvency` module. 

As this is an ERIC-routed service, call endpoints via the base url of: `http://api.chs.local:4001`. To check things are running, you might wish to try the healthcheck endpoint: `{base_url}/insolvency/healthcheck`

Development mode is available for this service in Docker CHS Development.

`./bin/chs-dev development enable insolvency-api`

## Running locally with Docker but without Docker CHS - Not recommended

Pull image from private CH registry by running docker pull 169942020521.dkr.ecr.eu-west-1.amazonaws.com/local/insolvency-api:latest command or run the following steps to build image locally:

1.  `export SSH_PRIVATE_KEY_PASSPHRASE='[your SSH key passhprase goes here]'` (optional, set only if SSH key is passphrase protected)
2.  `DOCKER_BUILDKIT=0 docker build --build-arg SSH_PRIVATE_KEY="$(cat ~/.ssh/id_rsa)" --build-arg SSH_PRIVATE_KEY_PASSPHRASE -t 169942020521.dkr.ecr.eu-west-1.amazonaws.com/local/insolvency-api .`
3.  `docker run 169942020521.dkr.ecr.eu-west-1.amazonaws.com/local/insolvency-api:latest`

However, this service has multiple dependencies e.g. on ERIC, Transactions API, Company Lookup, and EFS Submission API, so simply will not work in isolation.

## Running locally without Docker - Not recommended

1. Clone this repository: `go get github.com/companieshouse/insolvency-api`
1. Build the executable: `make build`

The same microservice dependecy considerations apply as with local Docker builds - this service is not intended to run in isolation.

## Configuration

| Variable                        | Default | Description             |
| :------------------------------ | :------ | :---------------------- |
| `BIND_ADDR`                     | `-`     | Insolvency API Port     |
| `MONGODB_URL`                   | `-`     | MongoDB URL             |
| `INSOLVENCY_MONGODB_DATABASE`   | `-`     | MongoDB database name   |
| `INSOLVENCY_MONGODB_COLLECTION` | `-`     | MongoDB collection name |

## Spec

When this service is live, specs will be available on the Companies House Developer Hub. As a courtesy during development, an OpenAPI 3 spec has been included in the `/apispec` folder, along with a description of how to use this API alongside the Transactions API (the Insolvency API is inseparable from that wider transactions-based model).


