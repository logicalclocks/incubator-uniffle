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
FROM golang:1.23 as builder

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


FROM debian:13-slim
# docker's ARG scope doc: https://docs.docker.com/engine/reference/builder/#scope, it's required to declear an arg in
# each stage.
ARG MODULE
WORKDIR /
# The packages below ship stale in debian:13-slim. Keep this list in sync with
# HWORKS-3166 and HWORKS-3167.
RUN apt-get update && \
    apt-get install -y --only-upgrade util-linux mount login libuuid1 libsmartcols1 \
        libmount1 liblastlog2-2 libblkid1 bsdutils openssl-provider-legacy libssl3t64 && \
    apt-get clean && rm -rf /var/lib/apt/lists/*
COPY --from=builder /workspace/${MODULE} .
ENV OPERATOR_BIN=${MODULE}
ENTRYPOINT ["/${OPERATOR_BIN}"]
