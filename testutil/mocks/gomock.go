package mocks

//go:generate mockgen -source=./../../x/contractmanager/types/expected_keepers.go -destination ./contractmanager/types/expected_keepers.go
//go:generate mockgen -source=./../../x/feerefunder/types/expected_keepers.go -destination ./feerefunder/types/keepers.go
//go:generate mockgen -source=./../../x/interchainqueries/types/verify.go -destination ./interchainqueries/keeper/verify.go
//go:generate mockgen -source=./../../x/interchainqueries/types/expected_keepers.go -destination ./interchainqueries/types/expected_keepers.go
//go:generate mockgen -source=./../../x/interchaintxs/types/expected_keepers.go -destination ./interchaintxs/types/expected_keepers.go
//go:generate mockgen -source=./../../x/transfer/types/expected_keepers.go -destination ./transfer/types/expected_keepers.go
//go:generate mockgen -source=./../../x/feeburner/types/expected_keepers.go -destination ./feeburner/types/expected_keepers.go
//go:generate mockgen -source=./../../x/cron/types/expected_keepers.go -destination ./cron/types/expected_keepers.go
//go:generate mockgen -source=./../../x/harpoon/types/expected_keepers.go -destination ./harpoon/types/expected_keepers.go
//go:generate mockgen -source=./../../x/revenue/types/expected_keepers.go -destination ./revenue/types/expected_keepers.go
