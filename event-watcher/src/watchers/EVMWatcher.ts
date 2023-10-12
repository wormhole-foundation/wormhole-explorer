import { Implementation__factory } from '@certusone/wormhole-sdk/lib/cjs/ethers-contracts/factories/Implementation__factory';
import { EVMChainName, coalesceChainId } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { Log } from '@ethersproject/abstract-provider';

import { BigNumber } from 'ethers';
import { AXIOS_CONFIG_JSON, NETWORK_CONTRACTS, NETWORK_RPCS_BY_CHAIN } from '../consts';
import { WHTransaction, VaasByBlock, WHTransferRedeemed } from '../databases/types';
import BaseWatcher from './BaseWatcher';
import {
  makeBlockKey,
  makeVaaKey,
  makeWHRedeemedTransaction,
  makeWHTransaction,
} from '../databases/utils';
import { makeSerializedVAA } from './utils';

export const wormholeInterface = Implementation__factory.createInterface();
// This is the hash for topic[0] of the core contract event LogMessagePublished
// https://github.com/wormhole-foundation/wormhole/blob/main/ethereum/contracts/Implementation.sol#L12
export const LOG_MESSAGE_PUBLISHED_TOPIC =
  '0x6eb224fb001ed210e379b335e35efe88672a8ce935d981a6896b27ffdf52a3b2';
// This is the hash for topic[0] of the token bridge contract event TransferRedeemed
// https://github.com/wormhole-foundation/wormhole/blob/99d01324b80d2e86d0e5b8ea832f9cf9d4119fcd/ethereum/contracts/bridge/Bridge.sol#L29
export const TRANSFER_REDEEMED_TOPIC =
  '0xcaf280c8cfeba144da67230d9b009c8f868a75bac9a528fa0474be1ba317c169';

export type BlockTag = 'finalized' | 'safe' | 'latest';
export type Block = {
  hash: string;
  number: number;
  timestamp: number;
};
export type ErrorBlock = {
  code: number; //6969,
  message: string; //'Error: No response received from RPC endpoint in 60s'
};

export class EVMWatcher extends BaseWatcher {
  finalizedBlockTag: BlockTag;
  lastTimestamp: number;
  latestFinalizedBlockNumber: number;

  constructor(chain: EVMChainName, finalizedBlockTag: BlockTag = 'latest') {
    super(chain);
    this.lastTimestamp = 0;
    this.latestFinalizedBlockNumber = 0;
    this.finalizedBlockTag = finalizedBlockTag;
    if (['acala', 'karura'].includes(chain)) {
      this.maximumBatchSize = 50;
    }
  }

  async getBlock(blockNumberOrTag: number | BlockTag): Promise<Block> {
    const rpc = NETWORK_RPCS_BY_CHAIN[this.chain];
    if (!rpc) {
      throw new Error(`${this.chain} RPC is not defined!`);
    }
    let result = (
      await this.http.post(
        rpc,
        [
          {
            jsonrpc: '2.0',
            id: 1,
            method: 'eth_getBlockByNumber',
            params: [
              typeof blockNumberOrTag === 'number'
                ? `0x${blockNumberOrTag.toString(16)}`
                : blockNumberOrTag,
              false,
            ],
          },
        ],
        AXIOS_CONFIG_JSON,
      )
    )?.data?.[0];
    if (result && result.result === null) {
      // Found null block
      if (
        typeof blockNumberOrTag === 'number' &&
        blockNumberOrTag < this.latestFinalizedBlockNumber - 1000
      ) {
        return {
          hash: '',
          number: BigNumber.from(blockNumberOrTag).toNumber(),
          timestamp: BigNumber.from(this.lastTimestamp).toNumber(),
        };
      }
    } else if (result && result.error && result.error.code === 6969) {
      return {
        hash: '',
        number: BigNumber.from(blockNumberOrTag).toNumber(),
        timestamp: BigNumber.from(this.lastTimestamp).toNumber(),
      };
    }
    result = result?.result;
    if (result && result.hash && result.number && result.timestamp) {
      // Convert to Ethers compatible type
      this.lastTimestamp = result.timestamp;
      return {
        hash: result.hash,
        number: BigNumber.from(result.number).toNumber(),
        timestamp: BigNumber.from(result.timestamp).toNumber(),
      };
    }
    throw new Error(
      `Unable to parse result of eth_getBlockByNumber for ${blockNumberOrTag} on ${rpc}`,
    );
  }

  async getBlocks(fromBlock: number, toBlock: number): Promise<Block[]> {
    const rpc = NETWORK_RPCS_BY_CHAIN[this.chain];
    if (!rpc) {
      throw new Error(`${this.chain} RPC is not defined!`);
    }
    const reqs: any[] = [];
    for (let blockNumber = fromBlock; blockNumber <= toBlock; blockNumber++) {
      reqs.push({
        jsonrpc: '2.0',
        id: (blockNumber - fromBlock).toString(),
        method: 'eth_getBlockByNumber',
        params: [`0x${blockNumber.toString(16)}`, false],
      });
    }
    const results = (await this.http.post(rpc, reqs, AXIOS_CONFIG_JSON))?.data;
    if (results && results.length) {
      // Convert to Ethers compatible type
      return results.map(
        (response: undefined | { result?: Block; error?: ErrorBlock }, idx: number) => {
          // Karura is getting 6969 errors for some blocks, so we'll just return empty blocks for those instead of throwing an error.
          // We take the timestamp from the previous block, which is not ideal but should be fine.
          if (
            (response &&
              response.result === null &&
              fromBlock + idx < this.latestFinalizedBlockNumber - 1000) ||
            (response?.error && response.error?.code && response.error.code === 6969)
          ) {
            return {
              hash: '',
              number: BigNumber.from(fromBlock + idx).toNumber(),
              timestamp: BigNumber.from(this.lastTimestamp).toNumber(),
            };
          }
          if (
            response?.result &&
            response.result?.hash &&
            response.result.number &&
            response.result.timestamp
          ) {
            this.lastTimestamp = response.result.timestamp;
            return {
              hash: response.result.hash,
              number: BigNumber.from(response.result.number).toNumber(),
              timestamp: BigNumber.from(response.result.timestamp).toNumber(),
            };
          }
          // console.error(reqs[idx], response, idx);
          throw new Error(
            `Unable to parse result of eth_getBlockByNumber for ${fromBlock + idx} on ${rpc}`,
          );
        },
      );
    }
    throw new Error(
      `Unable to parse result of eth_getBlockByNumber for range ${fromBlock}-${toBlock} on ${rpc}`,
    );
  }

  async getLogs(
    fromBlock: number,
    toBlock: number,
    address: string,
    topics: string[],
  ): Promise<Log[]> {
    const rpc = NETWORK_RPCS_BY_CHAIN[this.chain];
    if (!rpc) {
      throw new Error(`${this.chain} RPC is not defined!`);
    }
    const result = (
      await this.http.post(
        rpc,
        [
          {
            jsonrpc: '2.0',
            id: 1,
            method: 'eth_getLogs',
            params: [
              {
                fromBlock: `0x${fromBlock.toString(16)}`,
                toBlock: `0x${toBlock.toString(16)}`,
                address,
                topics,
              },
            ],
          },
        ],
        AXIOS_CONFIG_JSON,
      )
    )?.data?.[0]?.result;
    if (result) {
      // Convert to Ethers compatible type
      return result.map((l: Log) => ({
        ...l,
        blockNumber: BigNumber.from(l.blockNumber).toNumber(),
        transactionIndex: BigNumber.from(l.transactionIndex).toNumber(),
        logIndex: BigNumber.from(l.logIndex).toNumber(),
      }));
    }
    throw new Error(`Unable to parse result of eth_getLogs for ${fromBlock}-${toBlock} on ${rpc}`);
  }

  async getFinalizedBlockNumber(): Promise<number> {
    this.logger.debug(`fetching block ${this.finalizedBlockTag}`);
    const block: Block = await this.getBlock(this.finalizedBlockTag);
    this.latestFinalizedBlockNumber = block.number;
    return block.number;
  }

  override async getMessagesForBlocks(fromBlock: number, toBlock: number): Promise<VaasByBlock> {
    const address = NETWORK_CONTRACTS[this.chain].core;
    if (!address) {
      throw new Error(`Core contract not defined for ${this.chain}`);
    }
    const logs = await this.getLogs(fromBlock, toBlock, address, [LOG_MESSAGE_PUBLISHED_TOPIC]);
    const timestampsByBlock: { [block: number]: string } = {};
    // fetch timestamps for each block
    const vaasByBlock: VaasByBlock = {};
    this.logger.debug(`fetching info for blocks ${fromBlock} to ${toBlock}`);
    const blocks = await this.getBlocks(fromBlock, toBlock);
    for (const block of blocks) {
      const timestamp = new Date(block.timestamp * 1000).toISOString();
      timestampsByBlock[block.number] = timestamp;
      vaasByBlock[makeBlockKey(block.number.toString(), timestamp)] = [];
    }
    this.logger.debug(`processing ${logs.length} logs`);
    for (const log of logs) {
      const blockNumber = log.blockNumber;
      const emitter = log.topics[1].slice(2);
      const {
        args: { sequence },
      } = wormholeInterface.parseLog(log);
      const vaaKey = makeVaaKey(log.transactionHash, this.chain, emitter, sequence.toString());
      const blockKey = makeBlockKey(blockNumber.toString(), timestampsByBlock[blockNumber]);
      vaasByBlock[blockKey] = [...(vaasByBlock[blockKey] || []), vaaKey];
    }
    return vaasByBlock;
  }

  override async getWhEvents(
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

    // We collect the blocks data here to avoid making multiple requests to the RPC
    const blocks = await this.getBlocks(fromBlock, toBlock);
    const timestampsByBlock = [];
    for (const block of blocks) {
      const timestamp = new Date(block.timestamp * 1000);
      timestampsByBlock[block.number] = timestamp;
    }

    const sortedWhTxs = (await this.getWhTxs(fromBlock, toBlock, timestampsByBlock))?.sort(
      (a, b) => {
        return a.eventLog.sequence - b.eventLog.sequence;
      },
    );
    const sortedRedeemedTxs = await this.getRedeemedTxs(fromBlock, toBlock, timestampsByBlock);
    const lastSequenceNumber = await this.getLastSequenceNumber(sortedWhTxs);

    whEvents.whTxs = sortedWhTxs;
    whEvents.redeemedTxs = sortedRedeemedTxs;
    whEvents.lastSequenceNumber = lastSequenceNumber;

    return whEvents;
  }

  async getWhTxs(
    fromBlock: number,
    toBlock: number,
    timestampsByBlock?: Record<number, Date>,
  ): Promise<WHTransaction[]> {
    const whTxs: WHTransaction[] = [];
    const address = NETWORK_CONTRACTS[this.chain].core;

    if (!address) {
      throw new Error(`Core contract not defined for ${this.chain}`);
    }

    const txLogs = await this.getLogs(fromBlock, toBlock, address, [LOG_MESSAGE_PUBLISHED_TOPIC]);

    this.logger.debug(`processing ${txLogs.length} txLogs`);
    for (const txLog of txLogs) {
      // console.log('txLog', txLog);
      // console.log('txLog::parseLog', wormholeInterface.parseLog(txLog));

      const { args } = wormholeInterface.parseLog(txLog);
      const { sequence, payload, nonce, consistencyLevel } = args || {};
      const blockNumber = txLog.blockNumber;
      const chainName = this.chain;
      const chainId = coalesceChainId(chainName);
      const emitter = txLog.topics[1].slice(2);
      const parseSequence = Number(sequence.toString());
      const txHash = txLog.transactionHash;
      const parsePayload = Buffer.from(payload).toString().slice(2);
      const timestamp = timestampsByBlock![blockNumber];

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

    return whTxs;
  }

  async getRedeemedTxs(
    fromBlock: number,
    toBlock: number,
    timestampsByBlock?: Record<number, Date>,
  ): Promise<WHTransferRedeemed[]> {
    const redeemedTxs: WHTransferRedeemed[] = [];
    const tokenBridgeAddress = NETWORK_CONTRACTS[this.chain].token_bridge;

    if (!tokenBridgeAddress) {
      throw new Error(`Token Bridge contract not defined for ${this.chain}`);
    }

    const transferRedeemedLogs = await this.getLogs(fromBlock, toBlock, tokenBridgeAddress, [
      TRANSFER_REDEEMED_TOPIC,
    ]);

    this.logger.debug(`processing ${transferRedeemedLogs.length} transferRedeemedLogs`);
    for (const transferRedeemedLog of transferRedeemedLogs) {
      const { blockNumber, transactionHash, topics } = transferRedeemedLog;
      const [, emitterChainId, emitterAddress, sequence] = topics || [];

      if (emitterChainId && emitterAddress && sequence) {
        const parsedEmitterChainId = Number(emitterChainId.toString());
        const parsedEmitterAddress = emitterAddress.slice(2);
        const parsedSequence = Number(sequence.toString());
        const parsedBlockNumber = Number(blockNumber).toString(16);
        const indexedAt = timestampsByBlock![blockNumber];

        const redeemedTx = await makeWHRedeemedTransaction({
          emitterChainId: parsedEmitterChainId,
          emitterAddress: parsedEmitterAddress,
          sequence: parsedSequence,
          blockNumber: parsedBlockNumber,
          txHash: transactionHash,
          indexedAt,
          from: '',
          to: '',
        });

        redeemedTxs.push(redeemedTx);
      }
    }

    return redeemedTxs;
  }
}
