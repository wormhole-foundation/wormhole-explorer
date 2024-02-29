import { MetadataRepository } from "../../../domain/repositories";
import { HttpClientError } from "../../errors/HttpClientError";
import { EvmTag } from "../../../domain/entities";
import winston from "../../log";
import {
  EvmJsonRPCBlockRepository,
  EvmJsonRPCBlockRepositoryCfg,
  ProviderPoolMap,
} from "./EvmJsonRPCBlockRepository";

const FINALIZED = "finalized";
const ETHEREUM = "ethereum";

export class ArbitrumEvmJsonRPCBlockRepository extends EvmJsonRPCBlockRepository {
  override readonly logger = winston.child({ module: "ArbitrumEvmJsonRPCBlockRepository" });
  private latestL2Finalized: number;
  private metadataRepo: MetadataRepository<PersistedBlock[]>;
  private latestEthTime: number;

  constructor(
    cfg: EvmJsonRPCBlockRepositoryCfg,
    pools: ProviderPoolMap,
    metadataRepo: MetadataRepository<any>
  ) {
    super(cfg, pools);
    this.metadataRepo = metadataRepo;
    this.latestL2Finalized = 0;
    this.latestEthTime = 0;
  }

  async getBlockHeight(chain: string, finality: EvmTag): Promise<bigint> {
    const metadataFileName = `arbitrum-${finality}`;
    const chainCfg = this.getCurrentChain(chain);
    let response: { result: BlockByNumberResult };

    try {
      // This gets the latest L2 block so we can get the associated L1 block number
      response = await this.getChainProvider(chain).post<typeof response>(
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

    // Parser the L1 block number and L2 block number for arbitrum response
    const associatedL1ArbBlock: number = parseInt(l1BlockNumber, 16);
    const l2BlockArbNumber: number = parseInt(l2Number, 16);

    const persistedBlocks: PersistedBlock[] = (await this.metadataRepo.get(metadataFileName)) ?? [];
    const auxPersistedBlocks = this.removeDuplicates(persistedBlocks);

    // Only update the persisted block list, if the L2 block number is newer
    this.saveAssociatedL1Block(auxPersistedBlocks, associatedL1ArbBlock, l2BlockArbNumber); // 100

    // Only check every 30 seconds
    const now = Date.now();
    if (now - this.latestEthTime < 30_000) {
      return BigInt(this.latestL2Finalized);
    }
    this.latestEthTime = now;

    // Get the latest finalized L1 block ethereum number
    const latestL1BlockEthNumber: bigint = await super.getBlockHeight(ETHEREUM, FINALIZED);

    // Search in the persisted list looking for finalized L2 block number
    this.searchFinalizedBlock(auxPersistedBlocks, latestL1BlockEthNumber);

    await this.metadataRepo.save(metadataFileName, [...auxPersistedBlocks]);

    this.logger.info(
      `[${chain}] Blocks status: [PersistedBlocksLength: ${auxPersistedBlocks?.length}][Latest l2 arbi: ${l2BlockArbNumber} {Latest l1 arbi: ${associatedL1ArbBlock} - Latest l1 eth: ${latestL1BlockEthNumber}}, Latest l2 processed: ${this.latestL2Finalized}]`
    );

    const latestL2FinalizedToBigInt = this.latestL2Finalized;
    return BigInt(latestL2FinalizedToBigInt);
  }

  private removeDuplicates(persistedBlocks: PersistedBlock[]): PersistedBlock[] {
    const uniqueObjects = new Set();

    return persistedBlocks?.filter((obj) => {
      const key = JSON.stringify(obj);
      return !uniqueObjects.has(key) && uniqueObjects.add(key);
    });
  }

  private saveAssociatedL1Block(
    auxPersistedBlocks: PersistedBlock[],
    associatedL1ArbBlock: number,
    l2BlockArbNumber: number
  ): void {
    const findAssociatedL1Block = auxPersistedBlocks.find(
      (block) => block.associatedL1ArbBlock == associatedL1ArbBlock
    )?.associatedL1ArbBlock;

    if (!findAssociatedL1Block || findAssociatedL1Block < l2BlockArbNumber) {
      auxPersistedBlocks.push({
        associatedL1ArbBlock,
        l2BlockArbNumber,
        latestL2Finalized: this.latestL2Finalized,
      });
    }
  }

  private searchFinalizedBlock(
    auxPersistedBlocks: PersistedBlock[],
    latestL1BlockEthNumber: bigint
  ): void {
    const latestL1BlockNumberToNumber = Number(latestL1BlockEthNumber);
    const previusLatestL2Finalized =
      auxPersistedBlocks[auxPersistedBlocks.length - 1]?.latestL2Finalized;

    for (let index = auxPersistedBlocks.length - 1; index >= 0; index--) {
      const associatedL1ArbBlock = auxPersistedBlocks[index].associatedL1ArbBlock;

      if (associatedL1ArbBlock <= latestL1BlockNumberToNumber) {
        const l2BlockArbNumber = auxPersistedBlocks[index].l2BlockArbNumber;
        this.latestL2Finalized = l2BlockArbNumber;
        auxPersistedBlocks.splice(index, 1);
      }
    }

    if (this.latestL2Finalized == 0 || this.latestL2Finalized == previusLatestL2Finalized) {
      const l2BlockArbNumber = auxPersistedBlocks[0].l2BlockArbNumber;
      this.latestL2Finalized = l2BlockArbNumber;
      auxPersistedBlocks.splice(0, 1);
    }
  }
}

type PersistedBlock = {
  associatedL1ArbBlock: number;
  l2BlockArbNumber: number;
  latestL2Finalized: number;
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
