# tx-tracker service

The purpose of this service is to fetch, for each VAA, the metadata of associated to the transaction that generated it.

## Data flow

This service is consumes new VAAs from an SQS queue.

For most chains, the VAA includes a `txHash` field that identifies the transaction that generated the VAA
(for the particular case of Solana and Aptos, the value of this field is more like an indirect pointer to the transaction).

The service uses the value of this `txHash` field to query the associated chain's node RPC service, obtaining the transaction's metadata (e.g.: the sender's address).

This data is persisted in MongoDB the `globalTransaction` collection, as in the `originTx` object.

## Retry logic

Sometimes, fetching tx metadata from a node fails, e.g.:
1. The VAA was emitted with a low consistency level and the tx was not confirmed.
2. The particular node we're querying is down or we've reached a request limit.

When this happens, the service will retry until both these conditions are met:
* At least 10 minutes have passed since the VAA was emitted (this mitigates problem 1)
* At least 4 retries have been made (this mitigates problem 2)

This service processes VAAs in a sequential order, there is no concurrency.

## Backfiller

In the `cmd/backfiller` directory, there is a backfiller program that can be used to:
* Fill gaps in processing (e.g. due to downtime)
* Reprocess old records (e.g. overwrite them, this can be useful to apply database model changes)

The backfiller processes VAAs between two input dates.

VAAs will be iterated in a descending timestamp order, that is, newer VAAs are processed first. The rationale behind this is that old VAAs are not as relevant as recent ones.