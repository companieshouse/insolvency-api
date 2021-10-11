# Insolvency API

[![GoDoc](https://godoc.org/github.com/companieshouse/insolvency-api?status.svg)](https://godoc.org/github.com/companieshouse/insolvency-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/companieshouse/insolvency-api)](https://goreportcard.com/report/github.com/companieshouse/insolvency-api)

API code and specification for submitting insolvency data to Companies House

## Requirements
In order to run this API locally you will need to install the following:

- [Go](https://golang.org/doc/install)
- [Git](https://git-scm.com/downloads)
- [MongoDB](https://www.mongodb.com/)

## Getting Started
1. Clone this repository: `go get github.com/companieshouse/insolvency-api`
1. Build the executable: `make build`

## Configuration
Variable                        | Default   | Description
:-------------------------------|:----------|:------------
`BIND_ADDR`                     |`-`        | Insolvency API Port
`MONGODB_URL`                   |`-`        | MongoDB URL
`INSOLVENCY_MONGODB_DATABASE`   |`-`        | MongoDB database name
`INSOLVENCY_MONGODB_COLLECTION` |`-`        | MongoDB collection name

## Swagger UI

Insolvency API specifications schema: [https://github.com/companieshouse/insolvency-api/blob/master/apispec/schema.yml]().

Run the following to view the spec in the browser:

    docker-compose up -d

Spec can be viewed at [http://localhost:8080](http://localhost:8080)
