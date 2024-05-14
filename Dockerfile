# syntax=docker/dockerfile:1

FROM golang:1.22-bullseye
RUN apt-get update && apt-get install -y jq xxd
EXPOSE 26656 26657 1317 9090
COPY --from=app . /opt/neutron
RUN cd /opt/neutron && make install-test-binary
WORKDIR /opt/neutron

HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 CMD \
    curl -f http://127.0.0.1:1317/cosmos/base/tendermint/v1beta1/blocks/1 >/dev/null 2>&1 || exit 1

CMD bash /opt/neutron/network/init.sh && \
    bash /opt/neutron/network/init-neutrond.sh && \
    bash /opt/neutron/network/start.sh