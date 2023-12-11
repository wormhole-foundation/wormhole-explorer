import { FileMetadataRepository } from "../FileMetadataRepository";
import { MetadataRepository } from "../../../domain/repositories";
import { HttpClientError } from "../../errors/HttpClientError";
import { HttpClient } from "../../rpc/http/HttpClient";
import { EvmTag } from "../../../domain/entities";
import winston from "../../log";
import {
  EvmJsonRPCBlockRepository,
  EvmJsonRPCBlockRepositoryCfg,
} from "./EvmJsonRPCBlockRepository";

const FINALIZED = "finalized";
const ETHEREUM = "ethereum";

export class ArbitrumEvmJsonRPCBlockRepository extends EvmJsonRPCBlockRepository {
  override readonly logger = winston.child({ module: "ArbitrumEvmJsonRPCBlockRepository" });
  private metadataRepo: MetadataRepository<PersistedBlock[]>;
  private latestL2Finalized = 0;

  constructor(
    cfg: EvmJsonRPCBlockRepositoryCfg,
    httpClient: HttpClient,
    metadataRepo: MetadataRepository<any>
  ) {
    super(cfg, httpClient);
    this.metadataRepo = metadataRepo;
  }

  async getBlockHeight(chain: string, finality: EvmTag): Promise<bigint> {
    const chainCfg = this.getCurrentChain(chain);
    let response: { result: BlockByNumberResult };

    try {
      // This gets the latest L2 block so we can get the associated L1 block number
      response = await this.httpClient.post<typeof response>(
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
    const l2Number = l2Logs.number;

    if (!l2Logs || !l1BlockNumber || !l2Number)
      throw new Error(`[getBlockHeight] Unable to parse result for latest block on ${chain}`);

    const associatedL1Block: number = parseInt(l1BlockNumber, 16);
    const l2BlockNumber: number = parseInt(l2Number, 16);

    const persistedBlocks: PersistedBlock[] | undefined = await this.metadataRepo.get(
      `arbitrum-${finality}`
    );
    const auxPersistedBlocks = this.removeDuplicates(persistedBlocks);

    // Only update the persisted block list, if the L2 block number is newer
    this.saveAssociatedL1Block(auxPersistedBlocks, associatedL1Block, l2BlockNumber);

    // Get the latest finalized L1 block number
    const latestL1BlockNumber: bigint = await super.getBlockHeight(ETHEREUM, FINALIZED);

    // Search in the persisted list looking for finalized L2 block number
    this.searchFinalizedBlock(auxPersistedBlocks, latestL1BlockNumber);

    await this.metadataRepo.save(`arbitrum-${finality}`, [...auxPersistedBlocks]);

    const latestL2FinalizedToBigInt = this.latestL2Finalized;
    return BigInt(latestL2FinalizedToBigInt);
  }

  private removeDuplicates(persistedBlocks: PersistedBlock[] | undefined): PersistedBlock[] {
    const uniqueObjects = new Set();

    return (
      persistedBlocks?.filter((obj) => {
        const key = JSON.stringify(obj);
        return !uniqueObjects.has(key) && uniqueObjects.add(key);
      }) ?? []
    );
  }

  private saveAssociatedL1Block(
    auxPersistedBlocks: PersistedBlock[],
    associatedL1Block: number,
    l2BlockNumber: number
  ): void {
    const findAssociatedL1Block = auxPersistedBlocks.find(
      (block) => block.associatedL1Block == associatedL1Block
    )?.associatedL1Block;

    if (!findAssociatedL1Block || findAssociatedL1Block < l2BlockNumber) {
      auxPersistedBlocks.push({ associatedL1Block, l2BlockNumber });
    }
  }

  private searchFinalizedBlock(
    auxPersistedBlocks: PersistedBlock[],
    latestL1BlockNumber: bigint
  ): void {
    const latestL1BlockNumberToNumber = Number(latestL1BlockNumber);

    for (let index = auxPersistedBlocks.length - 1; index >= 0; index--) {
      const associatedL1Block = auxPersistedBlocks[index].associatedL1Block;

      if (associatedL1Block <= latestL1BlockNumberToNumber) {
        const l2BlockNumber = auxPersistedBlocks[index].l2BlockNumber;
        this.latestL2Finalized = l2BlockNumber;
        auxPersistedBlocks.splice(index, 1);
      }
    }
  }
}

type PersistedBlock = {
  associatedL1Block: number;
  l2BlockNumber: number;
};

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