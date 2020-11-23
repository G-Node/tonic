# SERVICE BUILDER IMAGE
FROM golang:alpine AS binbuilder

# Build dependencies
RUN apk --no-cache --no-progress add gcc musl-dev git curl

# Download git-annex to builder image and extract
RUN mkdir /git-annex
RUN curl -Lo /git-annex/git-annex-standalone-amd64.tar.gz https://downloads.kitenet.net/git-annex/linux/current/git-annex-standalone-amd64.tar.gz
RUN cd /git-annex && tar -xzf git-annex-standalone-amd64.tar.gz && rm git-annex-standalone-amd64.tar.gz

RUN go version
COPY ./go.mod ./go.sum /tonic/
WORKDIR /tonic

# Service to compile can be defined as a build arg.
# Default is example.
ARG service=example

# download deps before bringing in the sources
RUN go mod download
COPY ./templates /tonic/templates
COPY ./utonics /tonic/utonics
COPY ./tonic /tonic/tonic
RUN go build -v -o ${service} ./utonics/${service}/

### ============================ ###

# RUNNER IMAGE
FROM alpine:latest

RUN apk --no-cache --no-progress add git openssh

WORKDIR /tonic

# Copy git-annex from builder image
COPY --from=binbuilder /git-annex /git-annex
ENV PATH="${PATH}:/git-annex/git-annex.linux"

# Service to compile can be defined as a build arg.
# Default is example.
ARG service=example

# Copy binary and resources into runner image
COPY --from=binbuilder /tonic/${service} /tonic/service
COPY ./assets /tonic/assets

ENTRYPOINT /tonic/service
EXPOSE 3000
