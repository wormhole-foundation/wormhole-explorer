import { EvmTag } from "../../../domain/entities";
import { HttpClientError } from "../../errors/HttpClientError";
import { HttpClient } from "../../rpc/http/HttpClient";
import {
  EvmJsonRPCBlockRepository,
  EvmJsonRPCBlockRepositoryCfg,
} from "./EvmJsonRPCBlockRepository";

const ETHEREUM = "ethereum";
const FINALIZED = "finalized";

export class ArbitrumEvmJsonRPCBlockRepository extends EvmJsonRPCBlockRepository {
  private latestL2Finalized = 0;
  private l1L2Map = new Map<number, number>();

  constructor(cfg: EvmJsonRPCBlockRepositoryCfg, httpClient: HttpClient) {
    super(cfg, httpClient);
  }

  async getBlockHeight(chain: string, finality: EvmTag): Promise<bigint> {
    const chainCfg = this.getCurrentChain(chain);
    let response: { result: BlockByNumberResult };

    try {
      // This gets the latest L2 block so we can get the associated L1 block number
      response = await this.getHttpClient().post<typeof response>(
        chainCfg.rpc.href,
        {
          jsonrpc: "2.0",
          id: 1,
          method: "eth_getBlockByNumber",
          params: [finality, false],
        },
        { timeout: chainCfg.timeout, retries: chainCfg.retries }
      );
    } catch (e: HttpClientError | any) {
      this.handleError(chain, e, "getBlockHeight", "eth_getBlockByNumber");
      throw e;
    }

    const l2Logs = response.result;
    const l1BlockNumber = l2Logs.l1BlockNumber;
    const l1Number = l2Logs.number;

    if (!l2Logs || !l1BlockNumber || !l1Number)
      throw new Error(`[getBlockHeight] Unable to parse result for latest block on ${chain}`);

    const associatedL1Block: number = parseInt(l1BlockNumber, 16);
    const l2BlockNumber: number = parseInt(l1Number, 16);

    // Only update the map, if the L2 block number is newer
    const inMapL2 = this.l1L2Map.get(associatedL1Block);
    if (!inMapL2 || inMapL2 < l2BlockNumber) {
      this.l1L2Map.set(associatedL1Block, l2BlockNumber);
    }

    // Get the latest finalized L1 block number
    const latestL1BlockNumber: bigint = await super.getBlockHeight(ETHEREUM, FINALIZED);
    const latestL1BlockNumberToNumber = Number(latestL1BlockNumber);

    // Walk the map looking for finalized L2 block number
    for (const [l1, l2] of this.l1L2Map) {
      if (l1 <= latestL1BlockNumberToNumber) {
        this.latestL2Finalized = l2;
        this.l1L2Map.delete(l1);
      }
    }

    const latestL2FinalizedToBigInt = this.latestL2Finalized;
    return BigInt(latestL2FinalizedToBigInt);
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
