module github.com/wormhole-foundation/wormhole-explorer/parser

go 1.19

require (
	github.com/aws/aws-sdk-go v1.44.161
	github.com/gofiber/fiber/v2 v2.40.1
	github.com/ipfs/go-log/v2 v2.5.1
	github.com/joho/godotenv v1.4.0 // Configuration environment
	github.com/pkg/errors v0.9.1
	github.com/sethvargo/go-envconfig v0.6.0 // Configuration environment
	github.com/stretchr/testify v1.8.1 // indirect; Testing
	github.com/wormhole-foundation/wormhole/sdk v0.0.0-20221118153622-cddfe74b6787
	go.mongodb.org/mongo-driver v1.11.0
	go.uber.org/zap v1.23.0
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/ethereum/go-ethereum v1.10.21 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.15.11 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.41.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/xdg-go/stringprep v1.0.3 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/goleak v1.1.12 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/text v0.4.0 // indirect
)

// Needed for cosmos-sdk based chains.  See
// https://github.com/cosmos/cosmos-sdk/issues/10925 for more details.
replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
