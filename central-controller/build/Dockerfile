# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2021 Intel Corporation
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

FROM golang:1.18-alpine3.16

RUN apk update \
    && apk add --no-cache make

WORKDIR /go/src/github.com/central-controller
COPY ./ ./
RUN cd src/scc && make all
RUN cd src/rsync && make all
RUN cd src/monitor && make all

FROM alpine:3.16.0

RUN echo "https://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories
RUN apk del busybox \
    && apk update \
    && apk add openssl=1.1.1q-r0 \
    && apk add busybox=1.35.0-r18 \
    && rm /usr/bin/ssl_client

WORKDIR /opt/scc
RUN addgroup -S scc && adduser -S -G scc scc
RUN chown scc:scc /opt/scc -R

COPY --chown=scc --from=0 /go/src/github.com/central-controller/src/scc/scc ./
COPY --chown=scc --from=0 /go/src/github.com/central-controller/src/scc/ref-schemas ./ref-schemas/
COPY --chown=scc --from=0 /go/src/github.com/central-controller/src/rsync/rsync ./
COPY --chown=scc --from=0 /go/src/github.com/central-controller/src/monitor/monitor ./
COPY --chown=scc --from=0 /go/src/github.com/central-controller/build/entrypoint ./
RUN chmod +x entrypoint

USER scc

ENTRYPOINT ["./entrypoint"]
