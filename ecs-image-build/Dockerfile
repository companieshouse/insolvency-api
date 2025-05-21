FROM 416670754337.dkr.ecr.eu-west-2.amazonaws.com/ci-golang-build-1.24:latest

WORKDIR /opt

COPY /app .
COPY docker_start.sh .

RUN chmod +x docker_start.sh

CMD ["./docker_start.sh"]
