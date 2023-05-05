FROM 169942020521.dkr.ecr.eu-west-1.amazonaws.com/base/golang:1.19-bullseye-builder
FROM 169942020521.dkr.ecr.eu-west-1.amazonaws.com/base/golang:debian11-runtime
CMD ["-bind-addr=:10092"]
EXPOSE 10092