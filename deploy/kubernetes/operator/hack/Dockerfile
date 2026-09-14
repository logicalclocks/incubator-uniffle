#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# Build the binary
FROM golang:1.25 as builder

ARG MODULE
WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY api/ api/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o ${MODULE} cmd/${MODULE}/main.go


# Ubuntu rather than debian:13-slim: the Debian base shipped perl-base, the
# util-linux family, ncurses and sqlite3, whose 3 Critical and ~50 High CVEs have
# no fixed version to upgrade to. None of them are reachable from a static
# CGO_ENABLED=0 binary, but they cannot be cleared without leaving the distro.
#
# 24.04 rather than 26.04, measured as built here: 10 findings against 29, and
# nothing at CVSS >= 7.6 against 26.04's rust-coreutils 7.9 - which has no fix
# and exists only because 26.04 replaced GNU coreutils with Rust uutils. 24.04
# also ships no /usr/bin/pebble, so there is nothing to delete afterwards.
FROM ubuntu:24.04
# docker's ARG scope doc: https://docs.docker.com/engine/reference/builder/#scope, it's required to declear an arg in
# each stage.
ARG MODULE
WORKDIR /
RUN apt-get update && apt-get upgrade -y && apt-get clean && rm -rf /var/lib/apt/lists/*
COPY --from=builder /workspace/${MODULE} .
ENV OPERATOR_BIN=${MODULE}
ENTRYPOINT ["/${OPERATOR_BIN}"]
