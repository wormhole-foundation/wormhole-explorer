import { CONTRACTS } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import axios from 'axios';
import { RPCS_BY_CHAIN } from '../consts';
import { VaasByBlock } from '../databases/types';
import { makeBlockKey, makeVaaKey } from '../databases/utils';
import { EventObjectsTypes, RawLogEvents } from './TerraExplorerWatcher';
import { Watcher } from './Watcher';

export class InjectiveExplorerWatcher extends Watcher {
  // Arbitrarily large since the code here is capable of pulling all logs from all via indexer pagination
  maximumBatchSize: number = 1_000_000;

  latestBlockTag: string;
  getBlockTag: string;
  hashTag: string;
  contractTag: string;
  rpc: string | undefined;
  latestBlockHeight: number;

  constructor() {
    super('injective');
    this.rpc = RPCS_BY_CHAIN[this.chain];
    if (!this.rpc) {
      throw new Error(`${this.chain} RPC is not defined!`);
    }
    this.latestBlockHeight = 0;
    this.latestBlockTag = 'api/explorer/v1/blocks'; // This returns a page of the latest blocks
    this.getBlockTag = 'api/explorer/v1/blocks/';
    this.hashTag = 'api/explorer/v1/txs/';
    this.contractTag = 'api/explorer/v1/contractTxs/';
  }

  async getFinalizedBlockNumber(): Promise<number> {
    const result: ExplorerBlocks = (await axios.get(`${this.rpc}/${this.latestBlockTag}`)).data;
    if (result && result.paging.total) {
      let blockHeight: number = result.paging.total;
      if (blockHeight !== this.latestBlockHeight) {
        this.latestBlockHeight = blockHeight;
        this.logger.info('blockHeight = ' + blockHeight);
      }
      return blockHeight;
    }
    throw new Error(`Unable to parse result of ${this.latestBlockTag} on ${this.rpc}`);
  }

  // retrieve blocks for token bridge contract.
  // should be core, but the explorer doesn't support it yet
  // use "to": as the pagination key
  // compare block height ("block_number":) with what is passed in.
  async getMessagesForBlocks(fromBlock: number, toBlock: number): Promise<VaasByBlock> {
    const coreAddress = CONTRACTS.MAINNET[this.chain].core;
    const address = CONTRACTS.MAINNET[this.chain].token_bridge;
    if (!address) {
      throw new Error(`Token Bridge contract not defined for ${this.chain}`);
    }
    this.logger.debug(`Token Bridge contract for ${this.chain} is ${address}`);
    let vaasByBlock: VaasByBlock = {};
    this.logger.info(`fetching info for blocks ${fromBlock} to ${toBlock}`);

    const limit: number = 50;
    let done: boolean = false;
    let skip: number = 0;
    let lastBlockInserted: number = 0;
    while (!done) {
      // This URL gets the paginated list of transactions for the token bridge contract
      let url: string = `${this.rpc}/${this.contractTag}${address}?skip=${skip}&limit=${limit}`;
      this.logger.debug(`Query string = ${url}`);
      const bulkTxnResult = (
        await axios.get<ContractTxnResult>(url, {
          headers: {
            'User-Agent': 'Mozilla/5.0',
          },
        })
      ).data;
      if (!bulkTxnResult) {
        throw new Error('bad bulkTxnResult');
      }
      skip = bulkTxnResult.paging.to;
      const bulkTxns: ContractTxnData[] = bulkTxnResult.data;
      if (!bulkTxns) {
        throw new Error('No transactions');
      }
      for (let i: number = 0; i < bulkTxns.length; ++i) {
        // Walk the transactions
        const txn: ContractTxnData = bulkTxns[i];
        const height: number = txn.block_number;
        if (height >= fromBlock && height <= toBlock) {
          // We only care about the transactions in the given block range
          this.logger.debug(`Found one: ${fromBlock}, ${height}, ${toBlock}`);
          const blockKey = makeBlockKey(
            txn.block_number.toString(),
            new Date(txn.block_unix_timestamp).toISOString()
          );
          vaasByBlock[blockKey] = [];
          lastBlockInserted = height;
          this.logger.debug(`lastBlockInserted = ${lastBlockInserted}`);
          let vaaKey: string = '';
          // Each txn has an array of raw_logs
          if (txn.logs) {
            const rawLogs: RawLogEvents[] = txn.logs;
            for (let j: number = 0; j < rawLogs.length; ++j) {
              const rawLog: RawLogEvents = rawLogs[j];
              const events: EventObjectsTypes[] = rawLog.events;
              if (!events) {
                this.logger.debug(
                  `No events in rawLog${j} for block ${height}, hash = ${txn.hash}`
                );
                continue;
              }
              for (let k: number = 0; k < events.length; k++) {
                const event: EventObjectsTypes = events[k];
                if (event.type === 'wasm') {
                  if (event.attributes) {
                    let attrs = event.attributes;
                    let emitter: string = '';
                    let sequence: string = '';
                    let coreContract: boolean = false;
                    // only care about _contract_address, message.sender and message.sequence
                    const numAttrs = attrs.length;
                    for (let l = 0; l < numAttrs; l++) {
                      const key = attrs[l].key;
                      if (key === 'message.sender') {
                        emitter = attrs[l].value;
                      } else if (key === 'message.sequence') {
                        sequence = attrs[l].value;
                      } else if (key === '_contract_address' || key === 'contract_address') {
                        let addr = attrs[l].value;
                        if (addr === coreAddress) {
                          coreContract = true;
                        }
                      }
                    }
                    if (coreContract && emitter !== '' && sequence !== '') {
                      vaaKey = makeVaaKey(txn.hash, this.chain, emitter, sequence);
                      this.logger.debug('blockKey: ' + blockKey);
                      this.logger.debug('Making vaaKey: ' + vaaKey);
                      vaasByBlock[blockKey] = [...(vaasByBlock[blockKey] || []), vaaKey];
                    }
                  }
                }
              }
            }
          }
        }
        if (height < fromBlock) {
          this.logger.debug('Breaking out due to height < fromBlock');
          done = true;
          break;
        }
      }
      if (bulkTxns.length < limit) {
        this.logger.debug('Breaking out due to ran out of txns.');
        done = true;
      }
    }
    if (lastBlockInserted < toBlock) {
      // Need to create something for the last requested block because it will
      // become the new starting point for subsequent calls.
      this.logger.debug(`Adding filler for block ${toBlock}`);
      const blkUrl = `${this.rpc}/${this.getBlockTag}${toBlock}`;
      this.logger.debug(`Query string for block = ${blkUrl}`);
      const result = (await axios.get<ExplorerBlock>(blkUrl)).data;
      if (!result) {
        throw new Error(`Unable to get block information for block ${toBlock}`);
      }
      const blockKey = makeBlockKey(
        result.data.height.toString(),
        new Date(result.data.timestamp).toISOString()
      );
      vaasByBlock[blockKey] = [];
    }
    return vaasByBlock;
  }
}

type ExplorerBlocks = {
  paging: { total: number; from: number; to: number };
  data: ExplorerBlocksData[];
};

type ExplorerBlock = {
  s: string;
  data: ExplorerBlocksData;
};

type ExplorerBlocksData = {
  height: number;
  proposer: string;
  moniker: string;
  block_hash: string;
  parent_hash: string;
  num_pre_commits: number;
  num_txs: number;
  timestamp: string;
};

type ContractTxnResult = {
  data: ContractTxnData[];
  paging: {
    from: number;
    to: number;
    total: number;
  };
};

type ContractTxnData = {
  block_number: number;
  block_timestamp: string;
  block_unix_timestamp: number;
  code: number;
  codespace: string;
  data: string;
  error_log: string;
  gas_fee: {
    amount: Coin[];
    gas_limit: number;
    granter: string;
    payer: string;
  };
  gas_used: number;
  gas_wanted: number;
  hash: string;
  id: string;
  info: string;
  logs?: RawLogEvents[];
  memo: string;
  // messages: [];
  // signatures: [];
  tx_number: number;
  tx_type: string;
};

type Coin = {
  denom: string;
  amount: string;
};
