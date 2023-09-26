import { ChainName } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN, sleep } from '../common';
import { z } from 'zod';
import { TIMEOUT } from '../consts';
import { DBOptionTypes, WHTransaction, VaasByBlock } from '../databases/types';
import { getLogger, WormholeLogger } from '../utils/logger';
import { SNSOptionTypes } from '../services/SNS/types';
import { WatcherImplementation } from './types';
import { env } from '../config';

const isDev = env.NODE_ENV !== 'production';
abstract class BaseWatcher implements WatcherImplementation {
  public logger: WormholeLogger;
  maximumBatchSize: number = 100;
  sns?: SNSOptionTypes;
  db?: DBOptionTypes;
  stopWatcher: boolean = false;

  constructor(public chain: ChainName) {
    this.logger = getLogger(chain);
  }

  setDB(db: DBOptionTypes) {
    this.db = db;
  }

  setServices(sns: SNSOptionTypes) {
    this.sns = sns;
  }

  getMessagesForBlocks(_fromBlock: number, _toBlock: number): Promise<VaasByBlock> {
    throw new Error('Method not implemented.');
  }

  abstract getFinalizedBlockNumber(): Promise<number>;
  abstract getWhTxs(fromBlock: number, toBlock: number): Promise<WHTransaction[]>;

  isValidVaaKey(_key: string): boolean {
    throw new Error('Method not implemented.');
  }

  isValidBlockKey(key: string): boolean {
    try {
      const [block, timestamp] = key.split('/');
      const initialBlock = z
        .number()
        .int()
        .parse(Number(INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN[this.chain]));
      return (
        z.number().int().parse(Number(block)) > initialBlock &&
        Date.parse(z.string().datetime().parse(timestamp)) < Date.now()
      );
    } catch (e) {
      return false;
    }
  }

  async stop() {
    this.stopWatcher = true;
  }

  async watch(): Promise<void> {
    let toBlock: number | null = null;
    let fromBlock: number | null = this.db
      ? await this.db?.getResumeBlockByChain(this.chain)
      : null;
    let retry = 0;

    while (true) {
      if (this.stopWatcher) {
        this.logger.info(`Stopping Watcher...`);
        break;
      }

      if (isDev) {
        // Delay for 1 second in dev mode to avoid rate limiting
        await new Promise((resolve) => setTimeout(resolve, 1000));
      }

      try {
        if (fromBlock !== null && toBlock !== null && fromBlock <= toBlock) {
          // fetch logs for the block range, inclusive of toBlock
          toBlock = Math.min(fromBlock + this.maximumBatchSize - 1, toBlock);

          try {
            this.logger.debug(`fetching messages from ${fromBlock} to ${toBlock}`);
            // Here we get all the vaa logs from LOG_MESSAGE_PUBLISHED_TOPIC
            const whTxs = await this.getWhTxs(fromBlock, toBlock);

            if (whTxs?.length > 0) {
              // Then store the vaa logs processed in db
              await this.db?.storeWhTxs(this.chain, whTxs);

              // Then publish the vaa logs processed in SNS
              await this.sns?.publishMessages(whTxs, true);
            }
            // Then store the latest processed block by Chain Id
            await this.db?.storeLatestProcessBlock(this.chain, toBlock);
          } catch (e) {
            this.logger.error(e);
          }

          fromBlock = toBlock + 1;
        }

        try {
          this.logger.debug('fetching finalized block');
          toBlock = await this.getFinalizedBlockNumber();
          if (fromBlock === null) {
            // handle first loop on a fresh chain without initial block set
            fromBlock = toBlock;
          }
          retry = 0;
          await sleep(TIMEOUT);
        } catch (e) {
          // skip attempting to fetch messages until getting the finalized block succeeds
          toBlock = null;
          this.logger.error(`error fetching finalized block`);
          throw e;
        }
      } catch (e) {
        retry++;
        this.logger.error(e);
        const backOffTimeoutMS = TIMEOUT * 2 ** retry;
        this.logger.warn(`backing off for ${backOffTimeoutMS}ms`);
        await sleep(backOffTimeoutMS);
      }
    }
  }
}

export default BaseWatcher;
