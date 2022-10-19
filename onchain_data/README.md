# Onchain data for custody monitoring

For each connected wormhole chain, query the custody contract and get the updated locked value for each token.
There are two modes:
(1) query a known list of tracked tokens aka allow list
(2) query either a scan site (if available) or covalent (also if available) to grab list of known tokens locked in the custody contract

Set env variable allowlist=true if using allow list found at: onchain_data/data/allowList.json
taken from: https://github.com/wormhole-foundation/wormhole/blob/dev.v2/event_database/cloud_functions/token-allowlist-mainnet.json
Has the added benefit of also including the coin gecko ids in order to mark positions to USD

```
MONGODB_URI=mongodb://root:example@localhost:27017/ allowlist=true node lib/getCustodyData.js
```

Currently Near, Algorand, Xpla, & Aptos are not supported. Feel free to add support ;)

The aggregated custody data gets pushed to the mongodb database - "onchain_data" into the table - "custody"

There are api endpoints that the server provides which powers the wormscan website.
