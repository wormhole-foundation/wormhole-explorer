import { ChainName } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN, sleep } from '../common';
import { z } from 'zod';
import { TIMEOUT } from '../consts';
import { VaaLog, VaasByBlock } from '../databases/types';
import {
  getResumeBlockByChain,
  storeVaaLogs,
  storeVaasByBlock,
  storeLatestProcessBlock,
} from '../databases/utils';
import { getLogger, WormholeLogger } from '../utils/logger';
import AwsSNS from '../services/SNS/AwsSNS';
import { SNSConfig, SNSInput } from '../services/SNS/types';

const config: SNSConfig = {
  region: process.env.AWS_SNS_REGION as string,
  subject: process.env.AWS_SNS_SUBJECT as string,
  topicArn: process.env.AWS_TOPIC_ARN as string,
  credentials: {
    accessKeyId: process.env.AWS_ACCESS_KEY_ID as string,
    secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY as string,
  },
};

export class Watcher {
  chain: ChainName;
  logger: WormholeLogger;
  maximumBatchSize: number = 100;
  SNSClient: AwsSNS;

  constructor(chain: ChainName) {
    this.chain = chain;
    this.logger = getLogger(chain);
    this.SNSClient = new AwsSNS(config);
  }

  async getFinalizedBlockNumber(): Promise<number> {
    throw new Error('Not Implemented');
  }

  async getMessagesForBlocks(fromBlock: number, toBlock: number): Promise<VaasByBlock> {
    throw new Error('Not Implemented');
  }

  async getVaaLogs(fromBlock: number, toBlock: number): Promise<VaaLog[]> {
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

          // const vaasByBlock = await this.getMessagesForBlocks(fromBlock, toBlock);
          // await storeVaasByBlock(this.chain, vaasByBlock);

          // Here we get all the vaa logs from LOG_MESSAGE_PUBLISHED_TOPIC
          // Then store the latest processed block by Chain Id
          try {
            const vaaLogs = await this.getVaaLogs(fromBlock, toBlock);
            if (vaaLogs?.length > 0) {
              await storeVaaLogs(this.chain, vaaLogs);
              const messages: SNSInput[] = vaaLogs.map((log) => ({
                message: JSON.stringify({ ...log }),
              }));
              this.SNSClient.publishMessages(messages);
            }
            await storeLatestProcessBlock(this.chain, toBlock);
          } catch (e) {
            this.logger.error(e);
          }

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
