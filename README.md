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

As this is an ERIC-routed service, call endpoints via the base url of: `http://api.chs.local:4001`. To check things are running, you might wish to try the healthcheck endpoint: `{base_url}/insolvency-api/healthcheck`

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

## Terraform ECS

### What does this code do?

The code present in this repository is used to define and deploy a dockerised container in AWS ECS.
This is done by calling a [module](https://github.com/companieshouse/terraform-modules/tree/main/aws/ecs) from terraform-modules. Application specific attributes are injected and the service is then deployed using Terraform via the CICD platform 'Concourse'.

Application specific attributes | Value                                | Description
:---------|:-----------------------------------------------------------------------------|:-----------
**ECS Cluster**        | filing-close                                    | ECS cluster (stack) the service belongs to
**Load balancer**      |{env}-chs-apichgovuk <br> {env}-chs-apichgovuk-private                                 | The load balancer that sits in front of the service
**Concourse pipeline**     |[Pipeline link](https://ci-platform.companieshouse.gov.uk/teams/team-development/pipelines/insolvency-api) <br> [Pipeline code](https://github.com/companieshouse/ci-pipelines/blob/master/pipelines/ssplatform/team-development/insolvency-api)                                  | Concourse pipeline link in shared services


### Contributing
- Please refer to the [ECS Development and Infrastructure Documentation](https://companieshouse.atlassian.net/wiki/spaces/DEVOPS/pages/4390649858/Copy+of+ECS+Development+and+Infrastructure+Documentation+Updated) for detailed information on the infrastructure being deployed.

### Testing
- Ensure the terraform runner local plan executes without issues. For information on terraform runners please see the [Terraform Runner Quickstart guide](https://companieshouse.atlassian.net/wiki/spaces/DEVOPS/pages/1694236886/Terraform+Runner+Quickstart).
- If you encounter any issues or have questions, reach out to the team on the **#platform** slack channel.

### Vault Configuration Updates
- Any secrets required for this service will be stored in Vault. For any updates to the Vault configuration, please consult with the **#platform** team and submit a workflow request.

### Useful Links
- [ECS service config dev repository](https://github.com/companieshouse/ecs-service-configs-dev)
- [ECS service config production repository](https://github.com/companieshouse/ecs-service-configs-production)
