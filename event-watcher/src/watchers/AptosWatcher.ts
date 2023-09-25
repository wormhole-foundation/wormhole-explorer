import { coalesceChainId, CONTRACTS } from '@certusone/wormhole-sdk/lib/cjs/utils';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN } from '../common';
import { AptosClient } from 'aptos';
import { z } from 'zod';
import { NETWORK_CONTRACTS, NETWORK_RPCS_BY_CHAIN } from '../consts';
import { makeVaaKey, makeWHTransaction } from '../databases/utils';
import { AptosEvent } from '../types/aptos';
import BaseWatcher from './BaseWatcher';
import { WHTransaction, VaasByBlock } from '../databases/types';
import { makeSerializedVAA } from './utils';

const APTOS_CORE_BRIDGE_ADDRESS = NETWORK_CONTRACTS.aptos.core;
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
    this.client = new AptosClient(NETWORK_RPCS_BY_CHAIN[this.chain]!);
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

  override async getMessagesForBlocks(
    fromSequence: number,
    toSequence: number,
  ): Promise<VaasByBlock> {
    const limit = toSequence - fromSequence + 1;
    const events: AptosEvent[] = (await this.client.getEventsByEventHandle(
      APTOS_CORE_BRIDGE_ADDRESS,
      APTOS_EVENT_HANDLE,
      APTOS_FIELD_NAME,
      { start: fromSequence, limit },
    )) as AptosEvent[];
    const vaasByBlock: VaasByBlock = {};
    await Promise.all(
      events.map(async ({ data, sequence_number, version }) => {
        const [block, transaction] = await Promise.all([
          this.client.getBlockByVersion(Number(version)),
          this.client.getTransactionByVersion(Number(version)),
        ]);
        const timestamp = new Date(Number(block.block_timestamp) / 1000).toISOString();
        const blockKey = [block.block_height, timestamp, sequence_number].join('/'); // use custom block key for now so we can include sequence number
        const emitter = data.sender.padStart(64, '0');
        const vaaKey = makeVaaKey(transaction.hash, this.chain, emitter, data.sequence);
        vaasByBlock[blockKey] = [...(vaasByBlock[blockKey] ?? []), vaaKey];
      }),
    );
    return vaasByBlock;
  }

  override async getWhTxs(fromSequence: number, toSequence: number): Promise<WHTransaction[]> {
    const whTxs: WHTransaction[] = [];

    const limit = toSequence - fromSequence + 1;
    const events: AptosEvent[] = (await this.client.getEventsByEventHandle(
      APTOS_CORE_BRIDGE_ADDRESS,
      APTOS_EVENT_HANDLE,
      APTOS_FIELD_NAME,
      { start: fromSequence, limit },
    )) as AptosEvent[];

    await Promise.all(
      events.map(async (event) => {
        const { data, sequence_number, version } = event;
        const { consistency_level, sender, payload, nonce, timestamp } = data;
        const [transaction] = await Promise.all([
          this.client.getTransactionByVersion(Number(version)),
        ]);

        // console.log({ event });
        // console.log({ transaction, data, sequence_number, version });
        // console.log('------------------');

        // We store `blockNumber` with the sequence number.
        const blockNumber = sequence_number;
        const chainName = this.chain;
        const chainId = coalesceChainId(chainName);
        const parsedEmitter = sender.padStart(64, '0');
        const parseSequence = Number(sequence_number);
        const txHash = transaction.hash;
        const parsedNonce = Number(nonce);

        const vaaSerialized = await makeSerializedVAA({
          timestamp: Number(timestamp),
          nonce: parsedNonce,
          emitterChain: chainId,
          emitterAddress: parsedEmitter,
          sequence: parseSequence,
          payloadAsHex: payload.slice(2),
          consistencyLevel: consistency_level,
        });

        const unsignedVaaBuffer = Buffer.from(vaaSerialized, 'hex');

        const whTx = await makeWHTransaction({
          eventLog: {
            emitterChain: chainId,
            emitterAddr: parsedEmitter,
            sequence: parseSequence,
            txHash,
            blockNumber: blockNumber,
            unsignedVaa: unsignedVaaBuffer,
            sender: parsedEmitter,
            indexedAt: Number(timestamp),
          },
        });

        whTxs.push(whTx);
      }),
    );

    return whTxs;
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
