# SPDX-License-Identifier: Apache-2.0
# Copyright 2024 Authors of API-Speculator

FROM golang:1.23 AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Required to embed build info into binary.
COPY .git /.git

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} make build

FROM scratch

WORKDIR /speculator

COPY --from=builder /app/bin/speculator /speculator/
COPY --from=builder /app/config/default.yaml /speculator/config/default.yaml

ENTRYPOINT ["/speculator/speculator"]
