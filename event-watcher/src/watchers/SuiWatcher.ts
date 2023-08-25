import { CHAIN_ID_SUI } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import {
  Checkpoint,
  JsonRpcClient,
  PaginatedEvents,
  SuiTransactionBlockResponse,
} from '@mysten/sui.js';
import { array } from 'superstruct';
import { RPCS_BY_CHAIN } from '../consts';
import BaseWatcher from './BaseWatcher';
import { makeBlockKey, makeVaaKey, makeVaaLog } from '../databases/utils';
import { VaaLog, VaasByBlock } from '../databases/types';

const SUI_EVENT_HANDLE = `0x5306f64e312b581766351c07af79c72fcb1cd25147157fdc2f8ad76de9a3fb6a::publish_message::WormholeMessage`;

type PublishMessageEvent = {
  consistency_level: number;
  nonce: number;
  payload: number[];
  sender: string;
  sequence: string;
  timestamp: string;
};

export class SuiWatcher extends BaseWatcher {
  client: JsonRpcClient;
  override maximumBatchSize: number = 100_000; // arbitrarily large as this pages back by events

  constructor() {
    super('sui');
    this.client = new JsonRpcClient(RPCS_BY_CHAIN[this.chain]!);
  }

  // TODO: this might break using numbers, the whole service needs a refactor to use BigInt
  override async getFinalizedBlockNumber(): Promise<number> {
    return Number(
      (await this.client.request('sui_getLatestCheckpointSequenceNumber', undefined)).result,
    );
  }

  // TODO: this might break using numbers, the whole service needs a refactor to use BigInt
  override async getMessagesForBlocks(
    fromCheckpoint: number,
    toCheckpoint: number,
  ): Promise<VaasByBlock> {
    this.logger.info(`fetching info for checkpoints ${fromCheckpoint} to ${toCheckpoint}`);
    const vaasByBlock: VaasByBlock = {};

    {
      // reserve empty slot for initial block so query is cataloged
      const fromCheckpointTimestamp = new Date(
        Number(
          (
            await this.client.requestWithType(
              'sui_getCheckpoint',
              { id: fromCheckpoint.toString() },
              Checkpoint,
            )
          ).timestampMs,
        ),
      ).toISOString();
      const fromBlockKey = makeBlockKey(fromCheckpoint.toString(), fromCheckpointTimestamp);
      vaasByBlock[fromBlockKey] = [];
    }

    let lastCheckpoint: null | number = null;
    let cursor: any = undefined;
    let hasNextPage = false;
    do {
      const response = await this.client.requestWithType(
        'suix_queryEvents',
        {
          query: { MoveEventType: SUI_EVENT_HANDLE },
          cursor,
          descending_order: true,
        },
        PaginatedEvents,
      );
      const digest = response.data.length
        ? response.data[response.data.length - 1].id.txDigest
        : null;
      lastCheckpoint = digest
        ? Number(
            (
              await this.client.requestWithType(
                'sui_getTransactionBlock',
                { digest },
                SuiTransactionBlockResponse,
              )
            ).checkpoint!,
          )
        : null;
      cursor = response.nextCursor;
      hasNextPage = response.hasNextPage;
      const txBlocks = await this.client.requestWithType(
        'sui_multiGetTransactionBlocks',
        { digests: response.data.map((e) => e.id.txDigest) },
        array(SuiTransactionBlockResponse),
      );
      const checkpointByTxDigest = txBlocks.reduce<Record<string, string | undefined>>(
        (value, { digest, checkpoint }) => {
          value[digest] = checkpoint;
          return value;
        },
        {},
      );
      for (const event of response.data) {
        const checkpoint = checkpointByTxDigest[event.id.txDigest];
        if (!checkpoint) continue;
        const checkpointNum = Number(checkpoint);
        if (checkpointNum < fromCheckpoint || checkpointNum > toCheckpoint) continue;
        const msg = event.parsedJson as PublishMessageEvent;
        const timestamp = new Date(Number(msg.timestamp) * 1000).toISOString();
        const vaaKey = makeVaaKey(
          event.id.txDigest,
          CHAIN_ID_SUI,
          msg.sender.slice(2),
          msg.sequence,
        );
        const blockKey = makeBlockKey(checkpoint, timestamp);
        vaasByBlock[blockKey] = [...(vaasByBlock[blockKey] || []), vaaKey];
      }
    } while (hasNextPage && lastCheckpoint && fromCheckpoint < lastCheckpoint);
    return vaasByBlock;
  }

  override async getVaaLogs(fromCheckpoint: number, toCheckpoint: number): Promise<VaaLog[]> {
    const vaaLogs: VaaLog[] = [];
    let lastCheckpoint: null | number = null;
    let cursor: any = undefined;
    let hasNextPage = false;

    this.logger.info(`fetching info for checkpoints ${fromCheckpoint} to ${toCheckpoint}`);

    do {
      const response = await this.client.requestWithType(
        'suix_queryEvents',
        {
          query: { MoveEventType: SUI_EVENT_HANDLE },
          cursor,
          descending_order: true,
        },
        PaginatedEvents,
      );

      const digest = response.data.length
        ? response.data[response.data.length - 1].id.txDigest
        : null;

      lastCheckpoint = digest
        ? Number(
            (
              await this.client.requestWithType(
                'sui_getTransactionBlock',
                { digest },
                SuiTransactionBlockResponse,
              )
            ).checkpoint!,
          )
        : null;

      cursor = response.nextCursor;
      hasNextPage = response.hasNextPage;

      const txBlocks = await this.client.requestWithType(
        'sui_multiGetTransactionBlocks',
        { digests: response.data.map((e) => e.id.txDigest) },
        array(SuiTransactionBlockResponse),
      );

      const checkpointByTxDigest = txBlocks.reduce<Record<string, string | undefined>>(
        (value, { digest, checkpoint }) => {
          value[digest] = checkpoint;
          return value;
        },
        {},
      );

      for (const event of response.data) {
        const checkpoint = checkpointByTxDigest[event.id.txDigest];
        if (!checkpoint) continue;

        const checkpointNum = Number(checkpoint);
        if (checkpointNum < fromCheckpoint || checkpointNum > toCheckpoint) continue;

        const msg = event.parsedJson as PublishMessageEvent;
        // We store `blockNumber` with the checkpoint number.
        const blockNumber = checkpoint;
        const chainName = this.chain;
        const emitter = msg.sender.slice(2);
        const parsePayload = Buffer.from(msg.payload).toString('hex');
        const sequence = msg.sequence;
        const sender = msg.sender;
        const txHash = event.id.txDigest;

        const vaaLog = makeVaaLog({
          chainName,
          emitter,
          sequence,
          txHash,
          sender,
          blockNumber,
          payload: parsePayload,
        });

        vaaLogs.push(vaaLog);
      }
    } while (hasNextPage && lastCheckpoint && fromCheckpoint < lastCheckpoint);

    return vaaLogs;
  }
}
