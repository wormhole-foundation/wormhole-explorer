import { CONTRACTS } from '@certusone/wormhole-sdk/lib/cjs/utils';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN } from '../common';
import { AptosClient } from 'aptos';
import { z } from 'zod';
import { RPCS_BY_CHAIN } from '../consts';
import { makeVaaKey } from '../databases/utils';
import { AptosEvent } from '../types/aptos';
import BaseWatcher from './BaseWatcher';
import { VaaLog } from '../databases/types';

const APTOS_CORE_BRIDGE_ADDRESS = CONTRACTS.MAINNET.aptos.core;
const APTOS_EVENT_HANDLE = `${APTOS_CORE_BRIDGE_ADDRESS}::state::WormholeMessageHandle`;
const APTOS_FIELD_NAME = 'event';

/**
 * NOTE: The Aptos watcher differs from other watchers in that it uses the event sequence number to
 * fetch Wormhole messages and therefore also stores sequence numbers instead of block numbers.
 */
export class AptosWatcher extends BaseWatcher {
  client: AptosClient;
  override maximumBatchSize: number = 25;

  constructor() {
    super('aptos');
    this.client = new AptosClient(RPCS_BY_CHAIN[this.chain]!);
  }

  override async getFinalizedBlockNumber(): Promise<number> {
    return Number(
      (
        await this.client.getEventsByEventHandle(
          APTOS_CORE_BRIDGE_ADDRESS,
          APTOS_EVENT_HANDLE,
          APTOS_FIELD_NAME,
          { limit: 1 },
        )
      )[0].sequence_number,
    );
  }

  // async getMessagesForBlocks(fromSequence: number, toSequence: number): Promise<VaasByBlock> {
  //   const limit = toSequence - fromSequence + 1;
  //   const events: AptosEvent[] = (await this.client.getEventsByEventHandle(
  //     APTOS_CORE_BRIDGE_ADDRESS,
  //     APTOS_EVENT_HANDLE,
  //     APTOS_FIELD_NAME,
  //     { start: fromSequence, limit }
  //   )) as AptosEvent[];
  //   const vaasByBlock: VaasByBlock = {};
  //   await Promise.all(
  //     events.map(async ({ data, sequence_number, version }) => {
  //       const [block, transaction] = await Promise.all([
  //         this.client.getBlockByVersion(Number(version)),
  //         this.client.getTransactionByVersion(Number(version)),
  //       ]);
  //       const timestamp = new Date(Number(block.block_timestamp) / 1000).toISOString();
  //       const blockKey = [block.block_height, timestamp, sequence_number].join('/'); // use custom block key for now so we can include sequence number
  //       const emitter = data.sender.padStart(64, '0');
  //       const vaaKey = makeVaaKey(transaction.hash, this.chain, emitter, data.sequence);
  //       vaasByBlock[blockKey] = [...(vaasByBlock[blockKey] ?? []), vaaKey];
  //     })
  //   );
  //   return vaasByBlock;
  // }

  override getVaaLogs(fromBlock: number, toBlock: number): Promise<VaaLog[]> {
    throw new Error('Not Implemented');
  }

  override isValidBlockKey(key: string) {
    try {
      const [block, timestamp, sequence] = key.split('/');
      const initialSequence = z
        .number()
        .int()
        .parse(Number(INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN.aptos));
      return (
        z.number().int().parse(Number(block)) > 1094390 && // initial deployment block
        Date.parse(z.string().datetime().parse(timestamp)) < Date.now() &&
        z.number().int().parse(Number(sequence)) >= initialSequence // initial deployment sequence
      );
    } catch (e) {
      return false;
    }
  }

  override isValidVaaKey(key: string) {
    try {
      const [txHash, vaaKey] = key.split(':');
      const [_, emitter, sequence] = vaaKey.split('/');
      return (
        /^0x[0-9a-fA-F]{64}$/.test(z.string().parse(txHash)) &&
        /^[0-9]{64}$/.test(z.string().parse(emitter)) &&
        z.number().int().parse(Number(sequence)) >= 0
      );
    } catch (e) {
      return false;
    }
  }
}
