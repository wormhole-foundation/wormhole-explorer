import { ChainName } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN, sleep } from '../common';
import { z } from 'zod';
import { DEFAULT_RPS, NETWORK_RPS_BY_CHAIN, TIMEOUT } from '../consts';
import { DBOptionTypes, WHTransaction, VaasByBlock, WHTransferRedeemed } from '../databases/types';
import { getLogger, WormholeLogger } from '../utils/logger';
import { SNSOptionTypes } from '../services/SNS/types';
import { WatcherImplementation } from './types';
import axios from 'axios';
import rateLimit, { type RateLimitedAxiosInstance } from 'axios-rate-limit';

abstract class BaseWatcher implements WatcherImplementation {
  public logger: WormholeLogger;
  maximumBatchSize: number = 100;
  sns?: SNSOptionTypes;
  db?: DBOptionTypes;
  stopWatcher: boolean = false;
  http: RateLimitedAxiosInstance;

  constructor(public chain: ChainName) {
    this.logger = getLogger(chain);

    const rps = NETWORK_RPS_BY_CHAIN[this.chain] || DEFAULT_RPS;
    this.http = rateLimit(axios.create(), { perMilliseconds: 1000, maxRequests: rps });
  }

  abstract getFinalizedBlockNumber(): Promise<number>;
  abstract getWhTxs(fromBlock: number, toBlock: number): Promise<WHTransaction[]>;
  abstract getRedeemedTxs(fromBlock: number, toBlock: number): Promise<WHTransferRedeemed[]>;

  setDB(db: DBOptionTypes) {
    this.db = db;
  }

  setServices(sns: SNSOptionTypes) {
    this.sns = sns;
  }

  getMessagesForBlocks(_fromBlock: number, _toBlock: number): Promise<VaasByBlock> {
    throw new Error('Method not implemented.');
  }

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

  async getLastSequenceNumber(whTxs: WHTransaction[]): Promise<number | null> {
    if (whTxs.length > 0) {
      return whTxs[whTxs.length - 1].eventLog.sequence;
    }

    return null;
  }

  async getWhEvents(
    fromBlock: number,
    toBlock: number,
  ): Promise<{
    whTxs: WHTransaction[];
    redeemedTxs: WHTransferRedeemed[];
    lastSequenceNumber: number | null;
  }> {
    const whEvents: {
      whTxs: WHTransaction[];
      redeemedTxs: WHTransferRedeemed[];
      lastSequenceNumber: number | null;
    } = {
      whTxs: [],
      redeemedTxs: [],
      lastSequenceNumber: null,
    };

    const sortedWhTxs = (await this.getWhTxs(fromBlock, toBlock))?.sort((a, b) => {
      return a.eventLog.sequence - b.eventLog.sequence;
    });
    const sortedRedeemedTxs = await this.getRedeemedTxs(fromBlock, toBlock);
    const lastSequenceNumber = await this.getLastSequenceNumber(sortedWhTxs);

    whEvents.whTxs = sortedWhTxs;
    whEvents.redeemedTxs = sortedRedeemedTxs;
    whEvents.lastSequenceNumber = lastSequenceNumber;

    return whEvents;
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

      try {
        if (fromBlock !== null && toBlock !== null && fromBlock <= toBlock) {
          // fetch logs for the block range, inclusive of toBlock
          toBlock = Math.min(fromBlock + this.maximumBatchSize - 1, toBlock);

          try {
            this.logger.debug(`fetching messages from ${fromBlock} to ${toBlock}`);
            // Events from:
            // whTxs: LOG_MESSAGE_PUBLISHED_TOPIC (Core Contract)
            // redeemedTxs: TRANSFER_REDEEMED_TOPIC (Token Bridge Contract)
            const { whTxs, redeemedTxs, lastSequenceNumber } = await this.getWhEvents(
              fromBlock,
              toBlock,
            );

            if (whTxs?.length > 0) {
              // Then store the wormhole txs logs processed in db
              await this.db?.storeWhTxs(this.chain, whTxs);

              // Then publish the wormhole txs logs processed in SNS
              await this.sns?.createMessages(whTxs, 'whTx', true);
            }

            if (redeemedTxs?.length > 0) {
              // Then store the redeemed transfers logs processed in db
              await this.db?.storeRedeemedTxs(this.chain, redeemedTxs);
            }

            // Then store the latest processed block by Chain Id
            await this.db?.storeLatestProcessBlock(this.chain, toBlock, lastSequenceNumber);
          } catch (e: unknown) {
            let message;
            if (e instanceof Error) {
              message = e.message;
            } else {
              message = e;
            }

            this.logger.error(message);
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
          this.logger.error(`Error fetching finalized block`);
          throw e;
        }
      } catch (e) {
        retry++;
        let message;
        if (e instanceof Error) {
          message = e.message;
        } else {
          message = e;
        }

        this.logger.error(message);
        const backOffTimeoutMS = TIMEOUT * 2 ** retry;
        this.logger.warn(`backing off for ${backOffTimeoutMS}ms`);
        await sleep(backOffTimeoutMS);
      }
    }
  }
}

export default BaseWatcher;
