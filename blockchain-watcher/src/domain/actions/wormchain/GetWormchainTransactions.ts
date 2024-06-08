import { WormchainRepository } from "../../repositories";
import { WormchainBlockLogs } from "../../entities/wormchain";
import { mapChain } from "../../../common/wormchain";
import winston from "winston";

const ATTRIBUTES_TYPES = ["wasm"];

export class GetWormchainTransactions {
  private readonly blockRepo: WormchainRepository;
  protected readonly logger: winston.Logger;

  constructor(blockRepo: WormchainRepository) {
    this.logger = winston.child({ module: "GetWormchainTransactions" });
    this.blockRepo = blockRepo;
  }

  async execute(
    range: Range,
    opts: { addresses: string[]; chainId: number }
  ): Promise<WormchainBlockLogs[]> {
    const fromBlock = range.fromBlock;
    const toBlock = range.toBlock;
    const chain = mapChain(opts.chainId);

    if (fromBlock > toBlock) {
      this.logger.info(
        `[${chain}][exec] Invalid range [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
      );
      return [];
    }

    const blockNumbers: Set<bigint> = new Set();
    for (let block = fromBlock; block <= toBlock; block++) {
      blockNumbers.add(block);
    }

    const cosmosTransactions = await this.blockRepo.getBlockTransactions(
      opts.chainId,
      blockNumbers,
      ATTRIBUTES_TYPES
    );

    this.logger.info(
      `[${chain}][exec] Got ${
        cosmosTransactions?.length
      } transactions to process for ${this.populateLog(opts, fromBlock, toBlock)}`
    );
    return cosmosTransactions;
  }

  private populateLog(opts: { addresses: string[] }, fromBlock: bigint, toBlock: bigint): string {
    return `[addresses:${opts.addresses}][blocks:${fromBlock} - ${toBlock}]`;
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};
