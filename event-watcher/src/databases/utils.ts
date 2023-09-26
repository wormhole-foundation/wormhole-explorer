import { ChainId, ChainName, coalesceChainId } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { MAX_UINT_64, padUint16, padUint64 } from '../common';
import JsonDB from './JsonDB';
import MongoDB from './MongoDB';
import { env } from '../config';
import { DBOptionTypes, WHTransaction } from './types';
import { checkIfDateIsInMilliseconds } from '../utils/date';

// Bigtable Message ID format
// chain/MAX_UINT64-block/emitter/sequence
// 00002/00000000000013140651/0000000000000000000000008ea8874192c8c715e620845f833f48f39b24e222/00000000000000000000

export const getDB = (): DBOptionTypes => {
  if (env.DB_SOURCE === 'mongo') return new MongoDB();
  return new JsonDB();
};

export function makeMessageId(
  chainId: number,
  block: string,
  emitter: string,
  sequence: string,
): string {
  return `${padUint16(chainId.toString())}/${padUint64(
    (BigInt(MAX_UINT_64) - BigInt(block)).toString(),
  )}/${emitter}/${padUint64(sequence)}`;
}

export function parseMessageId(id: string): {
  chain: number;
  block: number;
  emitter: string;
  sequence: bigint;
} {
  const [chain, inverseBlock, emitter, sequence] = id.split('/');
  return {
    chain: parseInt(chain),
    block: Number(BigInt(MAX_UINT_64) - BigInt(inverseBlock)),
    emitter,
    sequence: BigInt(sequence),
  };
}

// TODO: should this be a composite key or should the value become more complex
export const makeBlockKey = (block: string, timestamp: string): string => `${block}/${timestamp}`;

export const makeVaaKey = (
  transactionHash: string,
  chain: ChainId | ChainName,
  emitter: string,
  seq: string,
): string => `${transactionHash}:${coalesceChainId(chain)}/${emitter}/${seq}`;

export const makeWHTransaction = async ({
  eventLog,
}: Omit<WHTransaction, 'id' | 'status'>): Promise<WHTransaction> => {
  const { emitterChain, emitterAddr, sequence, indexedAt } = eventLog;
  const vaaId = `${emitterChain}/${emitterAddr}/${sequence}`;
  const WH_TX_STATUS = 'created';

  let parsedIndexedAt = indexedAt;

  if (!(parsedIndexedAt instanceof Date)) {
    if (!checkIfDateIsInMilliseconds(parsedIndexedAt)) {
      parsedIndexedAt = new Date(+parsedIndexedAt * 1000) as unknown as number;
    }
  }

  return {
    id: vaaId,
    eventLog: {
      ...eventLog,
      indexedAt: parsedIndexedAt,
      createdAt: new Date(),
      updatedAt: new Date(),
      revision: 1,
    },
    status: WH_TX_STATUS,
  };
};
