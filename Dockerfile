# syntax=docker/dockerfile:1

FROM rust:1.63-bullseye as hermes-builder
WORKDIR /app
RUN PLATFORM=`uname -a | awk '{print $(NF-1)}'` && \
    git clone https://github.com/informalsystems/hermes.git && \
    cd hermes && \
    git checkout 7defaf067dbe6f60588518ea1619f228d38ac48d && \
    cargo build --release --bin hermes

FROM golang:1.18-bullseye
EXPOSE 16657
EXPOSE 16656
EXPOSE 6060
EXPOSE 1316 
EXPOSE 8090 
EXPOSE 8091 
EXPOSE 8080 
EXPOSE 26656 
EXPOSE 26657 
EXPOSE 1317 
EXPOSE 9090 
EXPOSE 9091
EXPOSE 8081
ADD . /opt/neutron
RUN cd /opt/neutron && make install
RUN mkdir -p $HOME/.hermes/bin
COPY --from=hermes-builder /app/hermes/target/release/hermes $HOME/.hermes/bin/

ENV PATH="${HOME}/.hermes/bin:${PATH}"
WORKDIR /opt/neutron
HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 CMD \
    curl -f http://127.0.0.1:1317/blocks/1 >/dev/null 2>&1 || exit 1
CMD ./network/init.sh && \
    ./network/start.sh && \
    ./network/hermes/restore-keys.sh && \
    ./network/hermes/create-conn.sh && \
    ./network/hermes/start.sh
