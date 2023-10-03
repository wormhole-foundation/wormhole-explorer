import { ChainName, coalesceChainId } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { readFileSync, writeFileSync } from 'fs';
import { env } from '../config';
import BaseDB from './BaseDB';
import { WHTransaction, WHTransferRedeemed } from './types';

const ENCODING = 'utf8';
const WORMHOLE_TX_FILE: string = env.JSON_WH_TXS_FILE;
const GLOBAL_TX_FILE: string = env.JSON_GLOBAL_TXS_FILE;
const WORMHOLE_LAST_BLOCKS_FILE: string = env.JSON_LAST_BLOCKS_FILE;

export default class JsonDB extends BaseDB {
  wormholeTxFile: WHTransaction[] = [];
  redeemedTxFile: WHTransferRedeemed[] = [];

  constructor() {
    super('JsonDB');
    this.wormholeTxFile = [];
    this.redeemedTxFile = [];
    this.lastBlocksByChain = [];
    this.logger.info('Connecting...');
  }

  async connect(): Promise<void> {
    try {
      const whTxsFileRawData = readFileSync(WORMHOLE_TX_FILE, ENCODING);
      this.wormholeTxFile = whTxsFileRawData ? JSON.parse(whTxsFileRawData) : [];
      this.logger.info(`${WORMHOLE_TX_FILE} file ready`);
    } catch (e) {
      this.logger.warn(`${WORMHOLE_TX_FILE} file does not exists, creating new file`);
      this.wormholeTxFile = [];
    }

    try {
      const whRedeemedTxsFileRawData = readFileSync(GLOBAL_TX_FILE, ENCODING);
      this.redeemedTxFile = whRedeemedTxsFileRawData ? JSON.parse(whRedeemedTxsFileRawData) : [];
      this.logger.info(`${GLOBAL_TX_FILE} file ready`);
    } catch (e) {
      this.logger.warn(`${GLOBAL_TX_FILE} file does not exists, creating new file`);
      this.redeemedTxFile = [];
    }
  }

  async disconnect(): Promise<void> {
    this.logger.info('Disconnecting...');
    this.logger.info('Disconnected');
  }

  async isConnected() {
    return true;
  }

  async getLastBlocksProcessed(): Promise<void> {
    try {
      const lastBlocksByChain = readFileSync(WORMHOLE_LAST_BLOCKS_FILE, ENCODING);
      this.lastBlocksByChain = lastBlocksByChain ? JSON.parse(lastBlocksByChain) : [];
      this.logger.info(`${WORMHOLE_LAST_BLOCKS_FILE} file ready`);
    } catch (e) {
      this.logger.warn(`${WORMHOLE_LAST_BLOCKS_FILE} file does not exists, creating new file`);
      this.lastBlocksByChain = [];
    }
  }

  override async storeWhTxs(chainName: ChainName, whTxs: WHTransaction[]): Promise<void> {
    try {
      for (let i = 0; i < whTxs.length; i++) {
        let message = 'Insert Wormhole Transaction Event Log to JSON file';
        const currentWhTx = whTxs[i];
        const { id } = currentWhTx;

        currentWhTx.eventLog.unsignedVaa = Buffer.isBuffer(currentWhTx.eventLog.unsignedVaa)
          ? Buffer.from(currentWhTx.eventLog.unsignedVaa).toString('base64')
          : currentWhTx.eventLog.unsignedVaa;

        const whTxIndex = this.wormholeTxFile?.findIndex((whTx) => whTx.id === id.toString());

        if (whTxIndex >= 0) {
          const whTx = this.wormholeTxFile[whTxIndex];

          whTx.eventLog.updatedAt = new Date();
          whTx.eventLog.revision ? (whTx.eventLog.revision += 1) : (whTx.eventLog.revision = 1);

          message = 'Update Wormhole Transaction Event Log to JSON file';
        } else {
          this.wormholeTxFile.push(currentWhTx);
        }

        writeFileSync(WORMHOLE_TX_FILE, JSON.stringify(this.wormholeTxFile, null, 2), ENCODING);

        if (currentWhTx) {
          const { id, eventLog } = currentWhTx;
          const { blockNumber, txHash, emitterChain } = eventLog;

          this.logger.info({
            id,
            blockNumber,
            chainName,
            txHash,
            emitterChain,
            message,
          });
        }
      }
    } catch (e: unknown) {
      this.logger.error(`Error Upsert Wormhole Transaction Event Log: ${e}`);
    }
  }

  override async storeRedeemedTxs(
    chainName: ChainName,
    redeemedTxs: WHTransferRedeemed[],
  ): Promise<void> {
    // For JsonDB we are only pushing all the "redeemed" logs into GLOBAL_TX_FILE simulating a globalTransactions collection

    try {
      for (let i = 0; i < redeemedTxs.length; i++) {
        let message = 'Insert Wormhole Transfer Redeemed Event Log to JSON file';
        const currentRedeemedTx = redeemedTxs[i];
        const { id, destinationTx } = currentRedeemedTx;
        const { method, status } = destinationTx;

        const whTxIndex = this.wormholeTxFile?.findIndex((whTx) => whTx.id === id.toString());

        if (whTxIndex >= 0) {
          const whTx = this.wormholeTxFile[whTxIndex];

          whTx.status = status;
          whTx.eventLog.updatedAt = new Date();
          whTx.eventLog.revision ? (whTx.eventLog.revision += 1) : (whTx.eventLog.revision = 1);

          writeFileSync(WORMHOLE_TX_FILE, JSON.stringify(this.wormholeTxFile, null, 2), ENCODING);
        }

        const whRedeemedTxIndex = this.redeemedTxFile?.findIndex(
          (whRedeemedTx) => whRedeemedTx.id === id.toString(),
        );

        if (whRedeemedTxIndex >= 0) {
          const whRedeemedTx = this.redeemedTxFile[whRedeemedTxIndex];

          whRedeemedTx.destinationTx.method = method;
          whRedeemedTx.destinationTx.status = status;
          whRedeemedTx.destinationTx.updatedAt = new Date();
          whRedeemedTx.revision ? (whRedeemedTx.revision += 1) : (whRedeemedTx.revision = 1);

          message = 'Update Wormhole Transfer Redeemed Event Log to JSON file';
        } else {
          this.redeemedTxFile.push(currentRedeemedTx);
        }

        writeFileSync(GLOBAL_TX_FILE, JSON.stringify(this.redeemedTxFile, null, 2), ENCODING);

        if (currentRedeemedTx) {
          const { id, destinationTx } = currentRedeemedTx;
          const { chainId } = destinationTx;

          this.logger.info({
            id,
            chainId,
            chainName,
            message,
          });
        }
      }
    } catch (e: unknown) {
      this.logger.error(`Error Upsert Wormhole Transfer Redeemed Event Log: ${e}`);
    }
  }

  override async storeLatestProcessBlock(chain: ChainName, lastBlock: number): Promise<void> {
    const chainId = coalesceChainId(chain);
    const updatedLastBlocksByChain = [...this.lastBlocksByChain];
    const itemIndex = updatedLastBlocksByChain.findIndex((item) => {
      if ('id' in item) return item.id === chain;
      return false;
    });

    if (itemIndex >= 0) {
      updatedLastBlocksByChain[itemIndex] = {
        ...updatedLastBlocksByChain[itemIndex],
        blockNumber: lastBlock,
        updatedAt: new Date(),
      };
    } else {
      updatedLastBlocksByChain.push({
        id: chain,
        blockNumber: lastBlock,
        chainId,
        createdAt: new Date(),
        updatedAt: new Date(),
      });
    }

    this.lastBlocksByChain = updatedLastBlocksByChain;

    try {
      writeFileSync(
        WORMHOLE_LAST_BLOCKS_FILE,
        JSON.stringify(this.lastBlocksByChain, null, 2),
        ENCODING,
      );
    } catch (e: unknown) {
      this.logger.error(`Error Insert latest processed block: ${e}`);
    }
  }
}
