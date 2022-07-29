FROM golang:1.18
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

RUN curl https://sh.rustup.rs -sSf | sh -s -- -y &&  /root/.cargo/bin/cargo install --version 0.14.1 ibc-relayer-cli --bin hermes --locked 
ADD . /opt/neutron
RUN cd /opt/neutron && make build
ENV PATH="/root/.cargo/bin/:${PATH}"
WORKDIR /opt/neutron

CMD make init && hermes -c ./network/hermes/config.toml create channel --port-a transfer --port-b transfer test-1 connection-0 && make start-rly 

