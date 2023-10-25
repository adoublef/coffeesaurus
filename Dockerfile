# syntax=docker/dockerfile:1

ARG GO_VERSION=1.21-alpine

FROM golang:${GO_VERSION} AS base

WORKDIR /usr/src

COPY go.* .
RUN go mod download

COPY . .

FROM base AS test

RUN go test -v -cover -count 1 ./...

FROM base AS build

ARG SOURCE_CODE=./cmd/coffeesaurus/

# cgo needed for litefs
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w -extldflags '-static'" \
    -buildvcs=false \
    -tags osusergo,netgo \
    -o /usr/bin/a ${SOURCE_CODE}

FROM alpine AS deploy

WORKDIR /opt

ENV PORT=8081

# copy binary from build
COPY --from=build /usr/bin/a .

ENTRYPOINT [ "./a" ]