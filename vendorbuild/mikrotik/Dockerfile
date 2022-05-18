FROM golang:1.18 AS builder
WORKDIR /app
COPY . .
RUN go build -o pathvector

FROM debian:11-slim
MAINTAINER info@pathvector.io

RUN apt-get update && apt-get -y install cron

COPY --from=builder /app/pathvector ./

COPY ./vendorbuild/mikrotik/crontab /etc/cron.d/pathvector-cron
RUN chmod 0644 /etc/cron.d/pathvector-cron

RUN touch /var/log/cron.log
ENTRYPOINT cron && tail -f /var/log/cron.log
