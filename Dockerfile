FROM 169942020521.dkr.ecr.eu-west-2.amazonaws.com/base/golang:1.19-bullseye-builder AS builder

RUN /bin/go_build

FROM 169942020521.dkr.ecr.eu-west-2.amazonaws.com/base/golang:debian11-runtime

COPY --from=builder /build/out/app ./

CMD ["-bind-addr=:10092"]
EXPOSE 10092