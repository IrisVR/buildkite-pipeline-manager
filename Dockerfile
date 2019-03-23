FROM alpine:latest

RUN apk add --no-cache ca-certificates

ADD buildkite-pipeline-manager /buildkite-pipeline-manager

ENTRYPOINT "/buildkite-pipeline-manager"