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

ARG SOURCE=./cmd/coffeesaurus/

# cgo needed for litefs
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w -extldflags '-static'" \
    -buildvcs=false \
    -tags osusergo,netgo \
    -o /usr/bin/a ${SOURCE}

FROM alpine AS deploy

WORKDIR /opt

ARG LITEFS_CONFIG="litefs.yml"
ENV LITEFS_DIR="/litefs"
ENV INTERNAL_PORT=8080
ENV PORT=8081

# copy binary from build
COPY --from=build /usr/bin/a .

# install sqlite, ca-certificates, curl and fuse for litefs
RUN apk add --no-cache bash fuse3 sqlite ca-certificates curl

# prepar for litefs
COPY --from=flyio/litefs:0.5 /usr/local/bin/litefs /usr/local/bin/litefs
ADD litefs/${LITEFS_CONFIG} /etc/litefs.yml
RUN mkdir -p /data ${LITEFS_DIR}

ENTRYPOINT ["litefs", "mount"]