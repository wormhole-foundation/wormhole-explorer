import algosdk from 'algosdk';
import BaseWatcher from './BaseWatcher';
import { ALGORAND_INFO } from '../consts';
import { makeBlockKey, makeVaaKey, makeWHTransaction } from '../databases/utils';
import { WHTransaction, VaasByBlock } from '../databases/types';
import { makeSerializedVAA } from './utils';
import { coalesceChainId } from '@certusone/wormhole-sdk';

type Message = {
  txHash: string | null;
  emitter: string;
  sequence: number | string;
  blockNumber: number | string | null;
  payload: any;
  blockKey: string;
  vaaKey: string;
  timestamp: Date;
};

export class AlgorandWatcher extends BaseWatcher {
  // Arbitrarily large since the code here is capable of pulling all logs from all via indexer pagination
  override maximumBatchSize: number = 100_000;

  algodClient: algosdk.Algodv2;
  indexerClient: algosdk.Indexer;

  constructor() {
    super('algorand');

    if (!ALGORAND_INFO.algodServer) {
      throw new Error('ALGORAND_INFO.algodServer is not defined!');
    }

    this.algodClient = new algosdk.Algodv2(
      ALGORAND_INFO.algodToken,
      ALGORAND_INFO.algodServer,
      ALGORAND_INFO.algodPort,
    );
    this.indexerClient = new algosdk.Indexer(
      ALGORAND_INFO.token,
      ALGORAND_INFO.server,
      ALGORAND_INFO.port,
    );
  }

  override async getFinalizedBlockNumber(): Promise<number> {
    this.logger.debug(`fetching final block for ${this.chain}`);

    let status = await this.algodClient.status().do();
    return status['last-round'];
  }

  async getApplicationLogTransactionIds(fromBlock: number, toBlock: number): Promise<string[]> {
    // it is possible this may result in gaps if toBlock > response['current-round']
    // perhaps to avoid this, getFinalizedBlockNumber could use the indexer?
    let transactionIds: string[] = [];
    let nextToken: string | undefined;
    let numResults: number | undefined;
    const maxResults = 225; // determined through testing
    do {
      const request = this.indexerClient
        .lookupApplicationLogs(ALGORAND_INFO.appid)
        .minRound(fromBlock)
        .maxRound(toBlock);
      if (nextToken) {
        request.nextToken(nextToken);
      }
      const response = await request.do();

      transactionIds = [
        ...transactionIds,
        ...(response?.['log-data']?.map((l: any) => l.txid) || []),
      ];
      nextToken = response?.['next-token'];
      numResults = response?.['log-data']?.length;
    } while (nextToken && numResults && numResults >= maxResults);
    return transactionIds;
  }

  processTransaction(transaction: any, parentId?: string): Message[] {
    let messages: Message[] = [];
    if (
      transaction['tx-type'] !== 'pay' &&
      transaction['application-transaction']?.['application-id'] === ALGORAND_INFO.appid &&
      transaction.logs?.length === 1
    ) {
      const txHash = parentId || transaction.id;
      const emitter = Buffer.from(algosdk.decodeAddress(transaction.sender).publicKey).toString(
        'hex',
      );
      const sequence = BigInt(
        `0x${Buffer.from(transaction.logs[0], 'base64').toString('hex')}`,
      ).toString();

      const blockNumber = transaction['confirmed-round'].toString();
      const payload = transaction['application-transaction']?.['application-args'][1];

      messages.push({
        txHash,
        emitter,
        sequence,
        blockNumber,
        payload,
        blockKey: makeBlockKey(
          blockNumber,
          new Date(transaction['round-time'] * 1000).toISOString(),
        ),
        vaaKey: makeVaaKey(txHash, this.chain, emitter, sequence),
        timestamp: new Date(transaction['round-time'] * 1000),
      });
    }
    if (transaction['inner-txns']) {
      for (const innerTransaction of transaction['inner-txns']) {
        messages = [...messages, ...this.processTransaction(innerTransaction, transaction.id)];
      }
    }

    return messages;
  }

  override async getMessagesForBlocks(fromBlock: number, toBlock: number): Promise<VaasByBlock> {
    const txIds = await this.getApplicationLogTransactionIds(fromBlock, toBlock);
    const transactions = [];
    for (const txId of txIds) {
      const response = await this.indexerClient.searchForTransactions().txid(txId).do();
      if (response?.transactions?.[0]) {
        transactions.push(response.transactions[0]);
      }
    }
    let messages: Message[] = [];
    for (const transaction of transactions) {
      messages = [...messages, ...this.processTransaction(transaction)];
    }
    const vaasByBlock = messages.reduce((vaasByBlock, message) => {
      if (!vaasByBlock[message.blockKey]) {
        vaasByBlock[message.blockKey] = [];
      }
      vaasByBlock[message.blockKey].push(message.vaaKey);
      return vaasByBlock;
    }, {} as VaasByBlock);
    const toBlockInfo = await this.indexerClient.lookupBlock(toBlock).do();
    const toBlockTimestamp = new Date(toBlockInfo.timestamp * 1000).toISOString();
    const toBlockKey = makeBlockKey(toBlock.toString(), toBlockTimestamp);
    if (!vaasByBlock[toBlockKey]) {
      vaasByBlock[toBlockKey] = [];
    }
    return vaasByBlock;
  }

  override async getWhTxs(fromBlock: number, toBlock: number): Promise<WHTransaction[]> {
    const whTxs: WHTransaction[] = [];
    const transactions = [];
    const txIds = await this.getApplicationLogTransactionIds(fromBlock, toBlock);

    for (const txId of txIds) {
      const response = await this.indexerClient.searchForTransactions().txid(txId).do();
      if (response?.transactions?.[0]) {
        transactions.push(response.transactions[0]);
      }
    }
    let messages: Message[] = [];
    for (const transaction of transactions) {
      messages = [...messages, ...this.processTransaction(transaction)];
    }

    // console.log({ messages });
    // console.log('-----');

    for (const message of messages) {
      const { txHash, emitter, sequence, blockNumber, payload, timestamp } = message;
      const chainName = this.chain;
      const chainId = coalesceChainId(chainName);
      const parseSequence = Number(sequence);
      const parsePayload = Buffer.from(payload, 'base64').toString('hex');

      const vaaSerialized = await makeSerializedVAA({
        timestamp,
        nonce: 0, // https://developer.algorand.org/docs/get-details/ethereum_to_algorand/#nonces-validity-windows-and-leases
        emitterChain: chainId,
        emitterAddress: emitter,
        sequence: parseSequence,
        payloadAsHex: parsePayload,
        consistencyLevel: 0, // https://docs.wormhole.com/wormhole/blockchain-environments/consistency#algorand
      });
      const unsignedVaaBuffer = Buffer.from(vaaSerialized, 'hex');

      const whTx = await makeWHTransaction({
        eventLog: {
          emitterChain: chainId,
          emitterAddr: emitter,
          sequence: parseSequence,
          txHash: txHash!,
          blockNumber: blockNumber!,
          unsignedVaa: unsignedVaaBuffer,
          sender: emitter,
          indexedAt: timestamp,
        },
      });

      whTxs.push(whTx);
    }

    return whTxs;
  }
}
