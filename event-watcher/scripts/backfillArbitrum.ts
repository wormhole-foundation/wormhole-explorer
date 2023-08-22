import * as dotenv from 'dotenv';
dotenv.config();
import { ChainName, CONTRACTS } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import axios from 'axios';
import ora from 'ora';
import { getDB } from '../src/databases/utils';
import { AXIOS_CONFIG_JSON } from '../src/consts';
import { ArbitrumWatcher } from '../src/watchers/ArbitrumWatcher';
import { LOG_MESSAGE_PUBLISHED_TOPIC } from '../src/watchers/EVMWatcher';

// This script exists because the Arbitrum RPC node only supports a 10 block range which is super slow

(async () => {
  const db = getDB();
  const chain: ChainName = 'arbitrum';
  const endpoint = `https://api.arbiscan.io/api?module=logs&action=getLogs&address=${CONTRACTS.MAINNET.arbitrum.core}&topic0=${LOG_MESSAGE_PUBLISHED_TOPIC}&apikey=YourApiKeyToken`;

  // fetch all message publish logs for core bridge contract from explorer
  let log = ora('Fetching logs from Arbiscan...').start();
  const blockNumbers = (await axios.get(endpoint, AXIOS_CONFIG_JSON)).data.result.map((x: any) =>
    parseInt(x.blockNumber, 16),
  );
  log.succeed(`Fetched ${blockNumbers.length} logs from Arbiscan`);
  // use the watcher to fetch corresponding blocks
  log = ora('Fetching blocks...').start();
  const watcher = new ArbitrumWatcher();
  for (const blockNumber of blockNumbers) {
    log.text = `Fetching block ${blockNumber}`;
    const vaasByBlock = await watcher.getMessagesForBlocks(blockNumber, blockNumber);
    await db.storeVaasByBlock(chain, vaasByBlock);
  }
  log.succeed('Uploaded messages to db successfully');
})();
