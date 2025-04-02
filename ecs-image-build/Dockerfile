FROM 416670754337.dkr.ecr.eu-west-2.amazonaws.com/ci-golang-build-1.23:latest

WORKDIR /opt

COPY /app .
COPY docker_start.sh .
CMD ["./docker_start.sh"]