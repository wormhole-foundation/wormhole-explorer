import { AXIOS_CONFIG_JSON, NETWORK_RPCS_BY_CHAIN } from '../consts';
import { EVMWatcher } from './EVMWatcher';

export class ArbitrumWatcher extends EVMWatcher {
  rpc: string | undefined;
  evmWatcher: EVMWatcher;
  latestL2Finalized: number;
  l1L2Map: Map<number, number>;
  lastEthTime: number;

  constructor() {
    super('arbitrum');

    this.rpc = NETWORK_RPCS_BY_CHAIN[this.chain];
    if (!this.rpc) {
      throw new Error(`${this.chain} RPC is not defined!`);
    }
    this.evmWatcher = new EVMWatcher('ethereum', 'finalized');
    this.latestL2Finalized = 0;
    this.l1L2Map = new Map<number, number>();
    this.lastEthTime = 0;
    this.maximumBatchSize = 25;
  }

  override async getFinalizedBlockNumber(): Promise<number> {
    if (!this.rpc) {
      throw new Error(`${this.chain} RPC is not defined!`);
    }

    // This gets the latest L2 block so we can get the associated L1 block number
    const l1Result: BlockByNumberResult = (
      await this.http.post(
        this.rpc,
        [
          {
            jsonrpc: '2.0',
            id: 1,
            method: 'eth_getBlockByNumber',
            params: ['latest', false],
          },
        ],
        AXIOS_CONFIG_JSON,
      )
    )?.data?.[0]?.result;
    if (!l1Result || !l1Result.l1BlockNumber || !l1Result.number) {
      throw new Error(
        `Unable to parse result of ArbitrumWatcher::eth_getBlockByNumber for latest on ${this.rpc}`,
      );
    }
    const associatedL1: number = parseInt(l1Result.l1BlockNumber, 16);
    const l2BlkNum: number = parseInt(l1Result.number, 16);
    this.logger.debug(
      'getFinalizedBlockNumber() checking map L1Block: ' +
        associatedL1 +
        ' => L2Block: ' +
        l2BlkNum,
    );

    // Only update the map, if the L2 block number is newer
    const inMapL2 = this.l1L2Map.get(associatedL1);
    if (!inMapL2 || inMapL2 < l2BlkNum) {
      this.logger.debug(`Updating map with ${associatedL1} => ${l2BlkNum}`);
      this.l1L2Map.set(associatedL1, l2BlkNum);
    }

    // Only check every 30 seconds
    const now = Date.now();
    if (now - this.lastEthTime < 30_000) {
      return this.latestL2Finalized;
    }
    this.lastEthTime = now;

    // Get the latest finalized L1 block number
    const evmFinal = await this.evmWatcher.getFinalizedBlockNumber();
    this.logger.debug(`Finalized EVM block number = ${evmFinal}`);

    this.logger.debug('Size of map = ' + this.l1L2Map.size);
    // Walk the map looking for finalized L2 block number
    for (const [l1, l2] of this.l1L2Map) {
      if (l1 <= evmFinal) {
        this.latestL2Finalized = l2;
        this.logger.debug(`Removing key ${l1} from map`);
        this.l1L2Map.delete(l1);
      }
    }

    this.logger.debug(`LatestL2Finalized = ${this.latestL2Finalized}`);
    return this.latestL2Finalized;
  }

  // This function is only used in test code.
  getFirstMapEntry(): number[] {
    if (this.l1L2Map.size > 0) {
      for (const [l1, l2] of this.l1L2Map) {
        return [l1, l2];
      }
    }
    return [0, 0];
  }
}

type BlockByNumberResult = {
  baseFeePerGas: string;
  difficulty: string;
  extraData: string;
  gasLimit: string;
  gasUsed: string;
  hash: string;
  l1BlockNumber: string;
  logsBloom: string;
  miner: string;
  mixHash: string;
  nonce: string;
  number: string;
  parentHash: string;
  receiptsRoot: string;
  sendCount: string;
  sendRoot: string;
  sha3Uncles: string;
  size: string;
  stateRoot: string;
  timestamp: string;
  totalDifficulty: string;
  transactions: string[];
  transactionsRoot: string;
  uncles: string[];
};
