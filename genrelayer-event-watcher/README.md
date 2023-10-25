# Explorer Event Watcher

The purpose of this process is to watch all Wormhole connected blockchains for events in Wormhole ecosystem contracts, and then produce database records for these events, and other important associated data.

For now, only EVM chains are supported, but this should be readily extensible in the future.

## Installation

run npm ci in the root of the project
run npm dev in the root of the project

## Deployment

This process is meant to be deployed as a docker container. The dockerfile is located in the root of the project.

## Configuration

Look at the provided example .env file for configuration options.

A brief explanation of the environment variables:
RPCS : A json object of rpcs to connect to. The key is the wormhole chain ID, and the value is the rpc url. If an RPC is provided here, it is always used, regardless of the USE_DEFAULT_RPCS variable.
ENVIRONMENT : DEVNET , TESTNET, or MAINNET these are Network objects from the wormhole SDK
USE_DEFAULT_RPCS : if true, the RPCS according to the ENVIRONMENT will be used. If false, an exception is thrown if the RPCS environment variable is not set for a given chain. RPCS

Contract overrides take the form of :
ETHEREUM_DEVNET_WORMHOLE_RELAYER_ADDRESS=0x53855d4b64E9A3CF59A84bc768adA716B5536BC5
BSC_DEVNET_WORMHOLE_RELAYER_ADDRESS=0x53855d4b64E9A3CF59A84bc768adA716B5536BC5

If the contract override is supplied, it will be used, otherwise the default contract addresses from the wormhole SDK will be used instead.

## Usage & Modification

A Handler should be created and registered in the handlers directory for each event type that needs to be handled. Handlers are registered inside the src/index.ts file. These handlers are treated as listeners, and thus require 100% uptime to to ensure
no events are missed.
