import { coalesceChainId } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { decode } from 'bs58';
import { Provider, TypedError } from 'near-api-js/lib/providers';
import { BlockResult, ExecutionStatus } from 'near-api-js/lib/providers/provider';
import ora from 'ora';
import { z } from 'zod';
import { NETWORK_CONTRACTS, NETWORK_RPCS_BY_CHAIN } from '../consts';
import { WHTransaction, VaasByBlock, WHTransferRedeemed } from '../databases/types';
import { makeBlockKey, makeVaaKey, makeWHTransaction } from '../databases/utils';
import { EventLog } from '../types/near';
import { getNearProvider, isWormholePublishEventLog } from '../utils/near';
import BaseWatcher from './BaseWatcher';
import { makeSerializedVAA } from '../utils/serializeVAA';

export class NearWatcher extends BaseWatcher {
  provider: Provider | null = null;
  override maximumBatchSize: number = 10;

  constructor() {
    super('near');
  }

  override async getFinalizedBlockNumber(): Promise<number> {
    this.logger.debug(`fetching final block for ${this.chain}`);
    const provider = await this.getProvider();
    const block = await provider.block({ finality: 'final' });
    return block.header.height;
  }

  override async getMessagesForBlocks(fromBlock: number, toBlock: number): Promise<VaasByBlock> {
    // assume toBlock was retrieved from getFinalizedBlockNumber and is finalized
    this.logger.debug(`fetching info for blocks ${fromBlock} to ${toBlock}`);
    const provider = await this.getProvider();
    const blocks: BlockResult[] = [];
    let block: BlockResult | null = null;
    try {
      block = await provider.block({ blockId: toBlock });
      blocks.push(block);
      while (true) {
        // traverse backwards via block hashes: https://github.com/wormhole-foundation/wormhole-monitor/issues/35
        block = await provider.block({ blockId: block.header.prev_hash });
        if (block.header.height < fromBlock) break;
        blocks.push(block);
      }
    } catch (e) {
      if (e instanceof TypedError && e.type === 'HANDLER_ERROR') {
        const error = block
          ? `block ${block.header.prev_hash} is too old, use backfillNear for blocks before height ${block.header.height}`
          : `toBlock ${toBlock} is too old, use backfillNear for this range`; // starting block too old
        this.logger.error(error);
      } else {
        throw e;
      }
    }

    return await getMessagesFromBlockResults(provider, blocks);
  }

  override async getWhTxs(fromBlock: number, toBlock: number): Promise<WHTransaction[]> {
    // assume toBlock was retrieved from getFinalizedBlockNumber and is finalized
    this.logger.debug(`fetching info for blocks ${fromBlock} to ${toBlock}`);
    const provider = await this.getProvider();
    const blocks: BlockResult[] = [];
    let block: BlockResult | null = null;
    try {
      block = await provider.block({ blockId: toBlock });
      blocks.push(block);
      while (true) {
        // traverse backwards via block hashes: https://github.com/wormhole-foundation/wormhole-monitor/issues/35
        block = await provider.block({ blockId: block.header.prev_hash });
        if (block.header.height < fromBlock) break;
        blocks.push(block);
      }
    } catch (e) {
      if (e instanceof TypedError && e.type === 'HANDLER_ERROR') {
        const error = block
          ? `block ${block.header.prev_hash} is too old, use backfillNear for blocks before height ${block.header.height}`
          : `toBlock ${toBlock} is too old, use backfillNear for this range`; // starting block too old
        this.logger.error(error);
      } else {
        throw e;
      }
    }

    return await getWhTxsResults(provider, blocks);
  }

  override async getRedeemedTxs(
    _fromBlock: number,
    _toBlock: number
  ): Promise<WHTransferRedeemed[]> {
    return [];
  }

  async getProvider(): Promise<Provider> {
    return (this.provider = this.provider || (await getNearProvider(NETWORK_RPCS_BY_CHAIN.near!)));
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

export const getMessagesFromBlockResults = async (
  provider: Provider,
  blocks: BlockResult[],
  debug: boolean = false
): Promise<VaasByBlock> => {
  const vaasByBlock: VaasByBlock = {};
  let log: ora.Ora;
  if (debug) log = ora(`Fetching messages from ${blocks.length} blocks...`).start();
  for (let i = 0; i < blocks.length; i++) {
    if (debug) log!.text = `Fetching messages from block ${i + 1}/${blocks.length}...`;
    const { height, timestamp } = blocks[i].header;
    const blockKey = makeBlockKey(height.toString(), new Date(timestamp / 1_000_000).toISOString());
    vaasByBlock[blockKey] = [];

    const chunks = [];
    for (const chunk of blocks[i].chunks) {
      chunks.push(await provider.chunk(chunk.chunk_hash));
    }

    const transactions = chunks.flatMap(({ transactions }) => transactions);
    for (const tx of transactions) {
      const outcome = await provider.txStatus(tx.hash, NETWORK_CONTRACTS.near.core);
      const logs = outcome.receipts_outcome
        .filter(
          ({ outcome }) =>
            (outcome as any).executor_id === NETWORK_CONTRACTS.near.core &&
            (outcome.status as ExecutionStatus).SuccessValue
        )
        .flatMap(({ outcome }) => outcome.logs)
        .filter((log) => log.startsWith('EVENT_JSON:')) // https://nomicon.io/Standards/EventsFormat
        .map((log) => JSON.parse(log.slice(11)) as EventLog)
        .filter(isWormholePublishEventLog);
      for (const log of logs) {
        const vaaKey = makeVaaKey(tx.hash, 'near', log.emitter, log.seq.toString());
        vaasByBlock[blockKey] = [...vaasByBlock[blockKey], vaaKey];
      }
    }
  }

  if (debug) {
    const numMessages = Object.values(vaasByBlock).flat().length;
    log!.succeed(`Fetched ${numMessages} messages from ${blocks.length} blocks`);
  }

  return vaasByBlock;
};

export const getWhTxsResults = async (
  provider: Provider,
  blocks: BlockResult[],
  debug: boolean = false
): Promise<WHTransaction[]> => {
  const whTxs: WHTransaction[] = [];

  let log: ora.Ora;
  if (debug) log = ora(`Fetching messages from ${blocks.length} blocks...`).start();
  for (let i = 0; i < blocks.length; i++) {
    if (debug) log!.text = `Fetching messages from block ${i + 1}/${blocks.length}...`;
    const { height, timestamp } = blocks[i].header;
    const blockNumber = height.toString();

    const chunks = [];
    for (const chunk of blocks[i].chunks) {
      chunks.push(await provider.chunk(chunk.chunk_hash));
    }

    const transactions = chunks.flatMap(({ transactions }) => transactions);
    for (const tx of transactions) {
      const outcome = await provider.txStatus(tx.hash, NETWORK_CONTRACTS.near.core);
      const logs = outcome.receipts_outcome
        .filter(({ outcome }) => {
          return (
            (outcome as any).executor_id === NETWORK_CONTRACTS.near.core &&
            (outcome.status as ExecutionStatus).SuccessValue
          );
        })
        .flatMap(({ outcome }) => outcome.logs)
        .filter((log) => log.startsWith('EVENT_JSON:')) // https://nomicon.io/Standards/EventsFormat
        .map((log) => JSON.parse(log.slice(11)) as EventLog)
        .filter(isWormholePublishEventLog);

      for (const log of logs) {
        const { nonce, emitter, seq, data } = log;

        const chainName = 'near';
        const chainId = coalesceChainId(chainName);
        const parseSequence = seq;
        const txHash = tx.hash;
        const parsePayload = data;
        const timestampDate = new Date(Math.floor(timestamp / 1_000_000)); // nanoseconds to milliseconds

        const vaaSerialized = await makeSerializedVAA({
          timestamp: timestampDate,
          nonce,
          emitterChain: chainId,
          emitterAddress: emitter,
          sequence: parseSequence,
          payloadAsHex: parsePayload,
          consistencyLevel: 0, // https://docs.wormhole.com/wormhole/blockchain-environments/consistency
        });
        const unsignedVaaBuffer = Buffer.from(vaaSerialized, 'hex');

        const whTx = await makeWHTransaction({
          eventLog: {
            emitterChain: chainId,
            emitterAddr: emitter,
            sequence: parseSequence,
            txHash,
            blockNumber: blockNumber,
            unsignedVaa: unsignedVaaBuffer,
            sender: '', // sender is not coming from the event log
            indexedAt: timestampDate,
          },
        });

        whTxs.push(whTx);
      }
    }
  }

  return whTxs;
};
