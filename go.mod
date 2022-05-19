module github.com/lidofinance/interchain-adapter

go 1.16

require (
	github.com/CosmWasm/wasmd v0.27.0
	github.com/cosmos/cosmos-sdk v0.45.4
	github.com/cosmos/ibc-go/v3 v3.0.0
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/prometheus/client_golang v1.12.1 // indirect
	github.com/regen-network/cosmos-proto v0.3.1
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.4.0
	github.com/stretchr/testify v1.7.1
	//github.com/tendermint/spn v0.2.1-0.20220511154430-aeab7a5b2bc0
	github.com/tendermint/tendermint v0.34.19
	github.com/tendermint/tm-db v0.6.7
	google.golang.org/genproto v0.0.0-20220407144326-9054f6ed7bac
	google.golang.org/grpc v1.45.0
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
