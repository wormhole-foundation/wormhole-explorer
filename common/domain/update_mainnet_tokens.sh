#!/bin/sh

rm -rf generated_mainnet_tokens.go
wget https://github.com/wormhole-foundation/wormhole/raw/main/node/pkg/governor/generated_mainnet_tokens.go
sed -i 's/package\ governor/package\ domain/g' generated_mainnet_tokens.go
sed -i 's/tokenConfigEntry/TokenMetadata/g' generated_mainnet_tokens.go
sed -i 's/chain\:/TokenChain\:/g' generated_mainnet_tokens.go
sed -i 's/addr\:/TokenAddress\:/g' generated_mainnet_tokens.go
sed -i 's/symbol\:/Symbol\:/g' generated_mainnet_tokens.go
sed -i 's/coinGeckoId\:/CoingeckoID\:/g' generated_mainnet_tokens.go
sed -i 's/decimals\:/Decimals\:/g' generated_mainnet_tokens.go
sed -i 's/TokenMetadata{TokenChain/{TokenChain/g' generated_mainnet_tokens.go