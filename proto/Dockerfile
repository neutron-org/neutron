FROM ghcr.io/cosmos/proto-builder:0.11.6

USER root

RUN apk add --no-cache \
  nodejs \
  npm \
  git \
  make \
  python3 \
  jq \
  bash

RUN npm install -g swagger-combine
RUN npm install -g swagger-merger
RUN npm install -g swagger2openapi

COPY --from=golang:1.23-alpine /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"