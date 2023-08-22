import { CONTRACTS } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { decode } from 'bs58';
import { Provider, TypedError } from 'near-api-js/lib/providers';
import { BlockResult, ExecutionStatus } from 'near-api-js/lib/providers/provider';
import ora from 'ora';
import { z } from 'zod';
import { RPCS_BY_CHAIN } from '../consts';
import { VaaLog } from '../databases/types';
import { makeBlockKey, makeVaaKey } from '../databases/utils';
import { EventLog } from '../types/near';
import { getNearProvider, isWormholePublishEventLog } from '../utils/near';
import BaseWatcher from './BaseWatcher';

export class NearWatcher extends BaseWatcher {
  provider: Provider | null = null;

  constructor() {
    super('near');
  }

  override async getFinalizedBlockNumber(): Promise<number> {
    this.logger.info(`fetching final block for ${this.chain}`);
    const provider = await this.getProvider();
    const block = await provider.block({ finality: 'final' });
    return block.header.height;
  }

  // async getMessagesForBlocks(fromBlock: number, toBlock: number): Promise<VaasByBlock> {
  //   // assume toBlock was retrieved from getFinalizedBlockNumber and is finalized
  //   this.logger.info(`fetching info for blocks ${fromBlock} to ${toBlock}`);
  //   const provider = await this.getProvider();
  //   const blocks: BlockResult[] = [];
  //   let block: BlockResult | null = null;
  //   try {
  //     block = await provider.block({ blockId: toBlock });
  //     blocks.push(block);
  //     while (true) {
  //       // traverse backwards via block hashes: https://github.com/wormhole-foundation/wormhole-monitor/issues/35
  //       block = await provider.block({ blockId: block.header.prev_hash });
  //       if (block.header.height < fromBlock) break;
  //       blocks.push(block);
  //     }
  //   } catch (e) {
  //     if (e instanceof TypedError && e.type === 'HANDLER_ERROR') {
  //       const error = block
  //         ? `block ${block.header.prev_hash} is too old, use backfillNear for blocks before height ${block.header.height}`
  //         : `toBlock ${toBlock} is too old, use backfillNear for this range`; // starting block too old
  //       this.logger.error(error);
  //     } else {
  //       throw e;
  //     }
  //   }

  //   return getMessagesFromBlockResults(provider, blocks);
  // }

  override getVaaLogs(fromBlock: number, toBlock: number): Promise<VaaLog[]> {
    throw new Error('Not Implemented');
  }

  async getProvider(): Promise<Provider> {
    return (this.provider = this.provider || (await getNearProvider(RPCS_BY_CHAIN.near!)));
  }

  override isValidVaaKey(key: string) {
    try {
      const [txHash, vaaKey] = key.split(':');
      const txHashDecoded = Buffer.from(decode(txHash)).toString('hex');
      const [_, emitter, sequence] = vaaKey.split('/');
      return (
        /^[0-9a-fA-F]{64}$/.test(z.string().parse(txHashDecoded)) &&
        /^[0-9a-fA-F]{64}$/.test(z.string().parse(emitter)) &&
        z.number().int().parse(Number(sequence)) >= 0
      );
    } catch (e) {
      return false;
    }
  }
}

// export const getMessagesFromBlockResults = async (
//   provider: Provider,
//   blocks: BlockResult[],
//   debug: boolean = false
// ): Promise<VaasByBlock> => {
//   const vaasByBlock: VaasByBlock = {};
//   let log: ora.Ora;
//   if (debug) log = ora(`Fetching messages from ${blocks.length} blocks...`).start();
//   for (let i = 0; i < blocks.length; i++) {
//     if (debug) log!.text = `Fetching messages from block ${i + 1}/${blocks.length}...`;
//     const { height, timestamp } = blocks[i].header;
//     const blockKey = makeBlockKey(height.toString(), new Date(timestamp / 1_000_000).toISOString());
//     vaasByBlock[blockKey] = [];

//     const chunks = [];
//     for (const chunk of blocks[i].chunks) {
//       chunks.push(await provider.chunk(chunk.chunk_hash));
//     }

//     const transactions = chunks.flatMap(({ transactions }) => transactions);
//     for (const tx of transactions) {
//       const outcome = await provider.txStatus(tx.hash, CONTRACTS.MAINNET.near.core);
//       const logs = outcome.receipts_outcome
//         .filter(
//           ({ outcome }) =>
//             (outcome as any).executor_id === CONTRACTS.MAINNET.near.core &&
//             (outcome.status as ExecutionStatus).SuccessValue
//         )
//         .flatMap(({ outcome }) => outcome.logs)
//         .filter((log) => log.startsWith('EVENT_JSON:')) // https://nomicon.io/Standards/EventsFormat
//         .map((log) => JSON.parse(log.slice(11)) as EventLog)
//         .filter(isWormholePublishEventLog);
//       for (const log of logs) {
//         const vaaKey = makeVaaKey(tx.hash, 'near', log.emitter, log.seq.toString());
//         vaasByBlock[blockKey] = [...vaasByBlock[blockKey], vaaKey];
//       }
//     }
//   }

//   if (debug) {
//     const numMessages = Object.values(vaasByBlock).flat().length;
//     log!.succeed(`Fetched ${numMessages} messages from ${blocks.length} blocks`);
//   }

//   return vaasByBlock;
// };
