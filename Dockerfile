FROM 169942020521.dkr.ecr.eu-west-1.amazonaws.com/base/golang:1.16-alpine-builder AS builder

FROM 169942020521.dkr.ecr.eu-west-1.amazonaws.com/base/golang:1.16-alpine-runtime AS runtime

EXPOSE 3055

CMD ["-bind-addr=:3055"]
