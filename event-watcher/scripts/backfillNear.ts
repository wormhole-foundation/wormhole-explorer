import * as dotenv from 'dotenv';
dotenv.config();
import { ChainName, CONTRACTS } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN } from '../src/common';
import { BlockResult } from 'near-api-js/lib/providers/provider';
import ora from 'ora';
import { getDB } from '../src/databases/utils';
import { getNearProvider, getTransactionsByAccountId, NEAR_ARCHIVE_RPC } from '../src/utils/near';
import { getMessagesFromBlockResults } from '../src/watchers/NearWatcher';

// This script exists because NEAR RPC nodes do not support querying blocks older than 5 epochs
// (~2.5 days): https://docs.near.org/api/rpc/setup#querying-historical-data. This script fetches
// all transactions for the core bridge contract from the NEAR Explorer backend API and then uses
// the archival RPC node to backfill messages in the given range.
//
// Ensure `DB_SOURCE` and Bigtable environment variables are set to backfill Bigtable database.
// Otherwise, the script will backfill the local JSON database.

const BATCH_SIZE = 1000;

(async () => {
  const db = getDB();
  const chain: ChainName = 'near';
  const provider = await getNearProvider(NEAR_ARCHIVE_RPC);
  const fromBlock = Number(
    (await db.getLastBlockByChain(chain)) ?? INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN[chain] ?? 0,
  );

  // fetch all transactions for core bridge contract from explorer
  let log = ora('Fetching transactions from NEAR Explorer...').start();
  const toBlock = await provider.block({ finality: 'final' });
  const transactions = await getTransactionsByAccountId(
    CONTRACTS.MAINNET.near.core,
    BATCH_SIZE,
    toBlock.header.timestamp.toString().padEnd(19, '9'), // pad to nanoseconds
  );
  log.succeed(`Fetched ${transactions.length} transactions from NEAR Explorer`);

  // filter out transactions that precede last seen block
  const blocks: BlockResult[] = [];
  const blockHashes = [...new Set(transactions.map((tx) => tx.blockHash))]; // de-dup blocks
  log = ora('Fetching blocks...').start();
  for (let i = 0; i < blockHashes.length; i++) {
    log.text = `Fetching blocks... ${i + 1}/${blockHashes.length}`;
    const block = await provider.block({ blockId: blockHashes[i] });
    if (block.header.height > fromBlock && block.header.height <= toBlock.header.height) {
      blocks.push(block);
    }
  }

  log.succeed(`Fetched ${blocks.length} blocks`);
  const vaasByBlock = await getMessagesFromBlockResults(provider, blocks, true);
  await db.storeVaasByBlock(chain, vaasByBlock);
  log.succeed('Uploaded messages to db successfully');
})();
