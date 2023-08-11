import { ChainName } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import {
  INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN,
  sleep,
} from '../common';
import { z } from 'zod';
import { TIMEOUT } from '../consts';
import { VaasByBlock } from '../databases/types';
import { getResumeBlockByChain, storeVaasByBlock } from '../databases/utils';
import { getLogger, WormholeLogger } from '../utils/logger';

export class Watcher {
  chain: ChainName;
  logger: WormholeLogger;
  maximumBatchSize: number = 100;

  constructor(chain: ChainName) {
    this.chain = chain;
    this.logger = getLogger(chain);
  }

  async getFinalizedBlockNumber(): Promise<number> {
    throw new Error('Not Implemented');
  }

  async getMessagesForBlocks(fromBlock: number, toBlock: number): Promise<VaasByBlock> {
    throw new Error('Not Implemented');
  }

  isValidBlockKey(key: string) {
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

  isValidVaaKey(key: string): boolean {
    throw new Error('Not Implemented');
  }

  async watch(): Promise<void> {
    let toBlock: number | null = null;
    let fromBlock: number | null = await getResumeBlockByChain(this.chain);
    let retry = 0;
    while (true) {
      try {
        if (fromBlock !== null && toBlock !== null && fromBlock <= toBlock) {
          // fetch logs for the block range, inclusive of toBlock
          toBlock = Math.min(fromBlock + this.maximumBatchSize - 1, toBlock);
          this.logger.info(`fetching messages from ${fromBlock} to ${toBlock}`);
          const vaasByBlock = await this.getMessagesForBlocks(fromBlock, toBlock);
          await storeVaasByBlock(this.chain, vaasByBlock);
          fromBlock = toBlock + 1;
        }
        try {
          this.logger.info('fetching finalized block');
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
        const expoBacko = TIMEOUT * 2 ** retry;
        this.logger.warn(`backing off for ${expoBacko}ms`);
        await sleep(expoBacko);
      }
    }
  }
}
