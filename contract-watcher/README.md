# contract-watcher

This component is in charge of obtaining the redeemed vaa's to obtain information from the target chain. .

## Usage

### Service

```bash
contract-watcher service
```

### Backfiller
```bash
contract-watcher backfiller [flags]
```

#### Command-line arguments
- **--chain-name** *string*                chain name
- **--chain-url** *string*                 chain URL
- **--from** *uint*                        first block to be processed
- **--log-level** *string*                 log level (default "INFO")
- **--mongo-database** *string*            mongo database
- **--mongo-uri** *string*                 mongo connection
- **--network** *string*                   network (mainnet or testnet)
- **--page-size** *int*                    maximum number to process at one time (default 100)
- **--persist-blocks**                     persist processed blocks in storage
- **--rate-limit** *int*                   rate limit per second (default 3)
- **--to** *uint*                          last block to be processed (included)
