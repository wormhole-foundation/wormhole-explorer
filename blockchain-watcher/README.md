# Explorer Event Watcher

The purpose of this process is to watch all Wormhole connected blockchains for events in Wormhole ecosystem contracts, and then produce database records for these events, and other important associated data.

For now, only EVM chains are supported, but this should be readily extensible in the future.

## Installation

run npm ci in the root of the project
run npm dev in the root of the project

## Deployment

This process is meant to be deployed as a docker container. The dockerfile is located in the root of the project.

## Configuration

Configuration is loaded from files in `config` directory.
There is a default file, and then a file for each environment. The environment is set by the NODE_ENV environment variable.
If NODE_ENV is not set, the default file is used.

Some values may be overriden by using environment variables. See `config/custom-environment-variables.json` for a list of these variables.

```bash
$ NODE_ENV=staging LOG_LEVEL=debug npm run dev
```

## Usage & Modification

A Handler should be created and registered in the handlers directory for each event type that needs to be handled. Handlers are registered inside the src/index.ts file. These handlers are treated as listeners, and thus require 100% uptime to to ensure
no events are missed.
