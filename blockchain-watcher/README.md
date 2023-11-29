# Explorer Blockchain Watcher

The purpose of this process is to watch all Wormhole connected blockchains for events in Wormhole ecosystem contracts, and then produce database records for these events, and other important associated data.

## Installation

run npm ci in the root of the project
run npm dev in the root of the project

## Deployment

This process is meant to be deployed as a docker container. The dockerfile is located in the root of the project.
As of today, we have a one to one relationship from job to pod. So we statically define jobs as a config map that each pod then loads.
See `../deploy/blockchain-watcher/workers` for an example of how deployment works.

To deploy all workers for a given environment, run:

```bash
$  kubetpl render workers/* -i env/staging-testnet.env --syntax=go-template | kubectl apply -f -
```

## Configuration

Configuration is loaded from files in `config` directory.
There is a default file, and then a file for each environment. The environment is set by the NODE_ENV environment variable.
If NODE_ENV is not set, the default file is used.

Some values may be overriden by using environment variables. See `config/custom-environment-variables.json` for a list of these variables.

```bash
$ NODE_ENV=staging LOG_LEVEL=debug npm run dev
```

When running locally, you need to configure one or more jobs.
By default, jobs are read from `metadata-repo/jobs/jobs.json`.

Example:

```
{ "jobs": [
  {
    "id": "poll-log-message-published-ethereum",
    "chain": "ethereum",
    "source": {
      "action": "PollEvmLogs",
      "config": {
        "fromBlock": "10012499",
        "blockBatchSize": 100,
        "commitment": "latest",
        "interval": 15000,
        "addresses": ["0x706abc4E45D419950511e474C7B9Ed348A4a716c"],
        "chain": "ethereum",
        "topics": []
      }
    },
    "handlers": [
      {
        "action": "HandleEvmLogs",
        "target": "sns",
        "mapper": "evmLogMessagePublishedMapper",
        "config": {
          "abi": "event LogMessagePublished(address indexed sender, uint64 sequence, uint32 nonce, bytes payload, uint8 consistencyLevel)",
          "filter": {
            "addresses": ["0x706abc4E45D419950511e474C7B9Ed348A4a716c"],
            "topics": ["0x6eb224fb001ed210e379b335e35efe88672a8ce935d981a6896b27ffdf52a3b2"]
          }
        }
      }
    ]
  }
]}

```

## Usage & Modification

Currently, jobs are read and loaded based on a JSON file.
Each job has a source, and one or more handlers.
Each handler has an action, a mapper and a target. For example, you can choose to use PollEvmLogs as an action and HandleEvmLogs as a handler. Fot this handler you need to set a mapper like evmLogMessagePublishedMapper.
The target can be sns, or a fake one if dryRun is enabled.
