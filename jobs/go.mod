module github.com/wormhole-foundation/wormhole-explorer/jobs

go 1.19

require (
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/joho/godotenv v1.5.1
	github.com/sethvargo/go-envconfig v0.9.0
	github.com/shopspring/decimal v1.3.1
	github.com/wormhole-foundation/wormhole-explorer/common v0.0.0-20230713181709-0425a89e7533
	go.uber.org/zap v1.24.0
)

require (
	github.com/algorand/go-algorand-sdk v1.23.0 // indirect
	github.com/algorand/go-codec/codec v1.1.8 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cosmos/btcutil v1.0.5 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/ethereum/go-ethereum v1.10.21 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/holiman/uint256 v1.2.1 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/onsi/gomega v1.27.6 // indirect
	github.com/wormhole-foundation/wormhole/sdk v0.0.0-20230426150516-e695fad0bed8 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/sys v0.9.0 // indirect
)

replace github.com/wormhole-foundation/wormhole-explorer/common => ../common