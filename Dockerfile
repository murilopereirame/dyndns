FROM golang:1.26-alpine AS build

WORKDIR /app

COPY . ./
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /dyndns ./cmd

FROM alpine:latest

ARG PUID=1000
ARG PGID=1000

WORKDIR /app
COPY --from=build /dyndns ./

RUN addgroup -g "${PGID}" -S go \
  && adduser -u "${PUID}" -G go -S -D -h /home/go go

USER go:go

ENTRYPOINT [ "/app/dyndns" ]
