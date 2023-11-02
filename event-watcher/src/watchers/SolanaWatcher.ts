import { getPostedMessage } from '@certusone/wormhole-sdk/lib/cjs/solana/wormhole';
import {
  Commitment,
  ConfirmedSignatureInfo,
  Connection,
  PublicKey,
  SolanaJSONRPCError,
  VersionedBlockResponse,
} from '@solana/web3.js';
import { decode } from 'bs58';
import { z } from 'zod';
import { NETWORK_CONTRACTS, NETWORK_RPCS_BY_CHAIN } from '../consts';
import { WHTransaction, VaasByBlock, WHTransferRedeemed } from '../databases/types';
import { makeBlockKey, makeVaaKey, makeWHTransaction } from '../databases/utils';
import { isLegacyMessage, normalizeCompileInstruction } from '../utils/solana';
import BaseWatcher from './BaseWatcher';
import { makeSerializedVAA } from '../utils/serializeVAA';

const WORMHOLE_PROGRAM_ID = NETWORK_CONTRACTS.solana.core;
const COMMITMENT: Commitment = 'finalized';
const GET_SIGNATURES_LIMIT = 1000;

export class SolanaWatcher extends BaseWatcher {
  rpc: string;
  // this is set as a class field so we can modify it in tests
  getSignaturesLimit = GET_SIGNATURES_LIMIT;
  // The Solana watcher uses the `getSignaturesForAddress` RPC endpoint to fetch all transactions
  // containing Wormhole messages. This API takes in signatures and paginates based on number of
  // transactions returned. Since we don't know the number of transactions in advance, we use
  // a block range of 100K slots. Technically, batch size can be arbitrarily large since pagination
  // of the WH transactions within that range is handled internally below.
  override maximumBatchSize = 10_000;

  constructor() {
    super('solana');
    this.rpc = NETWORK_RPCS_BY_CHAIN.solana!;
  }

  override async getFinalizedBlockNumber(): Promise<number> {
    const connection = new Connection(this.rpc, COMMITMENT);
    return connection.getSlot();
  }

  override async getMessagesForBlocks(fromSlot: number, toSlot: number): Promise<VaasByBlock> {
    const connection = new Connection(this.rpc, COMMITMENT);
    // in the rare case of maximumBatchSize skipped blocks in a row,
    // you might hit this error due to the recursion below
    if (fromSlot > toSlot) throw new Error('solana: invalid block range');
    this.logger.debug(`fetching info for blocks ${fromSlot} to ${toSlot}`);
    const vaasByBlock: VaasByBlock = {};

    // identify block range by fetching signatures of the first and last transactions
    // getSignaturesForAddress walks backwards so fromSignature occurs after toSignature
    let toBlock: VersionedBlockResponse | null = null;
    try {
      toBlock = await connection.getBlock(toSlot, { maxSupportedTransactionVersion: 0 });
    } catch (e) {
      if (e instanceof SolanaJSONRPCError && (e.code === -32007 || e.code === -32009)) {
        // failed to get confirmed block: slot was skipped or missing in long-term storage
        return this.getMessagesForBlocks(fromSlot, toSlot - 1);
      } else {
        throw e;
      }
    }
    if (!toBlock || !toBlock.blockTime || toBlock.transactions.length === 0) {
      return this.getMessagesForBlocks(fromSlot, toSlot - 1);
    }
    const fromSignature =
      toBlock.transactions[toBlock.transactions.length - 1].transaction.signatures[0];

    let fromBlock: VersionedBlockResponse | null = null;
    try {
      fromBlock = await connection.getBlock(fromSlot, { maxSupportedTransactionVersion: 0 });
    } catch (e) {
      if (e instanceof SolanaJSONRPCError && (e.code === -32007 || e.code === -32009)) {
        // failed to get confirmed block: slot was skipped or missing in long-term storage
        return this.getMessagesForBlocks(fromSlot + 1, toSlot);
      } else {
        throw e;
      }
    }
    if (!fromBlock || !fromBlock.blockTime || fromBlock.transactions.length === 0) {
      return this.getMessagesForBlocks(fromSlot + 1, toSlot);
    }
    const toSignature = fromBlock.transactions[0].transaction.signatures[0];

    // get all core bridge signatures between fromTransaction and toTransaction
    let numSignatures = this.getSignaturesLimit;
    let currSignature: string | undefined = fromSignature;
    while (numSignatures === this.getSignaturesLimit) {
      const signatures: ConfirmedSignatureInfo[] = await connection.getSignaturesForAddress(
        new PublicKey(WORMHOLE_PROGRAM_ID),
        {
          before: currSignature,
          until: toSignature,
          limit: this.getSignaturesLimit,
        }
      );

      this.logger.debug(`processing ${signatures.length} transactions`);

      // In order to determine if a transaction has a Wormhole message, we normalize and iterate
      // through all instructions in the transaction. Only PostMessage instructions are relevant
      // when looking for messages. PostMessageUnreliable instructions are ignored because there
      // are no data availability guarantees (ie the associated message accounts may have been
      // reused, overwriting previous data). Then, the message account is the account given by
      // the second index in the instruction's account key indices. From here, we can fetch the
      // message data from the account and parse out the emitter and sequence.
      const results = await connection.getTransactions(
        signatures.map((s) => s.signature),
        {
          maxSupportedTransactionVersion: 0,
        }
      );
      if (results.length !== signatures.length) {
        throw new Error(`solana: failed to fetch tx for signatures`);
      }
      for (const res of results) {
        if (res?.meta?.err) {
          // skip errored txs
          continue;
        }
        if (!res || !res.blockTime) {
          throw new Error(
            `solana: failed to fetch tx for signature ${
              res?.transaction.signatures[0] || 'unknown'
            }`
          );
        }

        const message = res.transaction.message;
        const accountKeys = isLegacyMessage(message)
          ? message.accountKeys
          : message.staticAccountKeys;
        const programIdIndex = accountKeys.findIndex((i) => i.toBase58() === WORMHOLE_PROGRAM_ID);
        const instructions = message.compiledInstructions;
        const innerInstructions =
          res.meta?.innerInstructions?.flatMap((i) =>
            i.instructions.map(normalizeCompileInstruction)
          ) || [];
        const whInstructions = innerInstructions
          .concat(instructions)
          .filter((i) => i.programIdIndex === programIdIndex);
        for (const instruction of whInstructions) {
          // skip if not postMessage instruction
          const instructionId = instruction.data;
          if (instructionId[0] !== 0x01) continue;

          const accountId = accountKeys[instruction.accountKeyIndexes[1]];
          const {
            message: { emitterAddress, sequence },
          } = await getPostedMessage(connection, accountId.toBase58(), COMMITMENT);
          const blockKey = makeBlockKey(
            res.slot.toString(),
            new Date(res.blockTime * 1000).toISOString()
          );
          const vaaKey = makeVaaKey(
            res.transaction.signatures[0],
            this.chain,
            emitterAddress.toString('hex'),
            sequence.toString()
          );
          vaasByBlock[blockKey] = [...(vaasByBlock[blockKey] || []), vaaKey];
        }
      }

      numSignatures = signatures.length;
      currSignature = signatures.at(-1)?.signature;
    }

    // add last block for storeVaasByBlock
    const lastBlockKey = makeBlockKey(
      toSlot.toString(),
      new Date(toBlock.blockTime * 1000).toISOString()
    );
    return { [lastBlockKey]: [], ...vaasByBlock };
  }

  override async getWhTxs(fromSlot: number, toSlot: number): Promise<WHTransaction[]> {
    const whTxs: WHTransaction[] = [];
    const connection = new Connection(this.rpc, COMMITMENT);
    // in the rare case of maximumBatchSize skipped blocks in a row,
    // you might hit this error due to the recursion below
    if (fromSlot > toSlot) throw new Error('solana: invalid block range');
    this.logger.debug(`fetching info for blocks ${fromSlot} to ${toSlot}`);

    // identify block range by fetching signatures of the first and last transactions
    // getSignaturesForAddress walks backwards so fromSignature occurs after toSignature
    let toBlock: VersionedBlockResponse | null = null;
    try {
      toBlock = await connection.getBlock(toSlot, { maxSupportedTransactionVersion: 0 });
    } catch (e) {
      if (e instanceof SolanaJSONRPCError && (e.code === -32007 || e.code === -32009)) {
        // failed to get confirmed block: slot was skipped or missing in long-term storage
        return this.getWhTxs(fromSlot, toSlot - 1);
      } else {
        throw e;
      }
    }
    if (!toBlock || !toBlock.blockTime || toBlock.transactions.length === 0) {
      return this.getWhTxs(fromSlot, toSlot - 1);
    }
    const fromSignature =
      toBlock.transactions[toBlock.transactions.length - 1].transaction.signatures[0];

    let fromBlock: VersionedBlockResponse | null = null;
    try {
      fromBlock = await connection.getBlock(fromSlot, { maxSupportedTransactionVersion: 0 });
    } catch (e) {
      if (e instanceof SolanaJSONRPCError && (e.code === -32007 || e.code === -32009)) {
        // failed to get confirmed block: slot was skipped or missing in long-term storage
        return this.getWhTxs(fromSlot + 1, toSlot);
      } else {
        throw e;
      }
    }
    if (!fromBlock || !fromBlock.blockTime || fromBlock.transactions.length === 0) {
      return this.getWhTxs(fromSlot + 1, toSlot);
    }
    const toSignature = fromBlock.transactions[0].transaction.signatures[0];

    // get all core bridge signatures between fromTransaction and toTransaction
    let numSignatures = this.getSignaturesLimit;
    let currSignature: string | undefined = fromSignature;
    while (numSignatures === this.getSignaturesLimit) {
      const signatures: ConfirmedSignatureInfo[] = await connection.getSignaturesForAddress(
        new PublicKey(WORMHOLE_PROGRAM_ID),
        {
          before: currSignature,
          until: toSignature,
          limit: this.getSignaturesLimit,
        }
      );

      this.logger.debug(`processing ${signatures.length} transactions`);

      // In order to determine if a transaction has a Wormhole message, we normalize and iterate
      // through all instructions in the transaction. Only PostMessage instructions are relevant
      // when looking for messages. PostMessageUnreliable instructions are ignored because there
      // are no data availability guarantees (ie the associated message accounts may have been
      // reused, overwriting previous data). Then, the message account is the account given by
      // the second index in the instruction's account key indices. From here, we can fetch the
      // message data from the account and parse out the emitter and sequence.
      const results = await connection.getTransactions(
        signatures.map((s) => s.signature),
        {
          maxSupportedTransactionVersion: 0,
        }
      );
      if (results.length !== signatures.length) {
        throw new Error(`solana: failed to fetch tx for signatures`);
      }
      for (const res of results) {
        if (res?.meta?.err) {
          // skip errored txs
          continue;
        }
        if (!res || !res.blockTime) {
          throw new Error(
            `solana: failed to fetch tx for signature ${
              res?.transaction.signatures[0] || 'unknown'
            }`
          );
        }

        const message = res.transaction.message;
        const accountKeys = isLegacyMessage(message)
          ? message.accountKeys
          : message.staticAccountKeys;
        const programIdIndex = accountKeys.findIndex((i) => i.toBase58() === WORMHOLE_PROGRAM_ID);
        const instructions = message.compiledInstructions;
        const innerInstructions =
          res.meta?.innerInstructions?.flatMap((i) =>
            i.instructions.map(normalizeCompileInstruction)
          ) || [];
        const whInstructions = innerInstructions
          .concat(instructions)
          .filter((i) => i.programIdIndex === programIdIndex);
        for (const instruction of whInstructions) {
          // skip if not postMessage instruction

          const instructionId = instruction.data;
          if (instructionId[0] !== 0x01) continue;

          const accountId = accountKeys[instruction.accountKeyIndexes[1]];
          const { message } = await getPostedMessage(connection, accountId.toBase58(), COMMITMENT);
          const {
            sequence,
            emitterAddress,
            emitterChain,
            submissionTime: timestamp,
            nonce,
            payload,
            consistencyLevel,
          } = message || {};

          // console.log('res', res);
          // console.log(
          //   'parseLog',
          //   await getPostedMessage(connection, accountId.toBase58(), COMMITMENT),
          // );
          // console.log('-----');

          // We store `blockNumber` with the slot number.
          const blockNumber = res.slot.toString();
          const chainId = emitterChain;
          const emitter = emitterAddress.toString('hex');
          const parsePayload = payload.toString('hex');
          const parseSequence = Number(sequence);
          const txHash = res.transaction.signatures[0];

          const vaaSerialized = await makeSerializedVAA({
            timestamp,
            nonce,
            emitterChain: chainId,
            emitterAddress: emitter,
            sequence: parseSequence,
            payloadAsHex: parsePayload,
            consistencyLevel,
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
              indexedAt: timestamp,
            },
          });

          whTxs.push(whTx);
        }
      }

      numSignatures = signatures.length;
      currSignature = signatures.at(-1)?.signature;
    }

    return whTxs;
  }

  override async getRedeemedTxs(
    _fromBlock: number,
    _toBlock: number
  ): Promise<WHTransferRedeemed[]> {
    return [];
  }

  override isValidVaaKey(key: string) {
    try {
      const [txHash, vaaKey] = key.split(':');
      const txHashDecoded = Buffer.from(decode(txHash)).toString('hex');
      const [, emitter, sequence] = vaaKey.split('/');
      return !!(
        /^[0-9a-fA-F]{128}$/.test(z.string().parse(txHashDecoded)) &&
        /^[0-9a-fA-F]{64}$/.test(z.string().parse(emitter)) &&
        z.number().int().parse(Number(sequence)) >= 0
      );
    } catch (e) {
      return false;
    }
  }
}
