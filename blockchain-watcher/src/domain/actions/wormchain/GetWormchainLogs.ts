import { WormchainRepository } from "../../repositories";
import { WormchainBlockLogs } from "../../entities/wormchain";
import winston from "winston";

const ATTRIBUTES_TYPES = ["wasm"];

export class GetWormchainLogs {
  private readonly blockRepo: WormchainRepository;
  protected readonly logger: winston.Logger;

  constructor(blockRepo: WormchainRepository) {
    this.logger = winston.child({ module: "GetWormchainLogs" });
    this.blockRepo = blockRepo;
  }

  async execute(
    range: Range,
    opts: { addresses: string[]; chain: string }
  ): Promise<WormchainBlockLogs[]> {
    const fromBlock = range.fromBlock;
    const toBlock = range.toBlock;

    const collectWormchainLogs: WormchainBlockLogs[] = [];

    if (fromBlock > toBlock) {
      this.logger.info(
        `[wormchain][exec] Invalid range [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
      );
      return [];
    }
    this.logger.info(
      `[wormchain][exec] Processing blocks [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
    );

    for (let blockNumber = fromBlock; blockNumber <= toBlock; blockNumber++) {
      const wormchainLogs = await this.blockRepo.getBlockLogs(
        opts.chain,
        blockNumber,
        ATTRIBUTES_TYPES
      );

      if (wormchainLogs && wormchainLogs.transactions && wormchainLogs.transactions.length > 0) {
        collectWormchainLogs.push(wormchainLogs);
      }
    }

    this.logger.info(
      `[wormchain][exec] Got ${
        collectWormchainLogs?.length
      } transactions to process for ${this.populateLog(opts, fromBlock, toBlock)}`
    );
    return collectWormchainLogs;
  }

  private populateLog(opts: { addresses: string[] }, fromBlock: bigint, toBlock: bigint): string {
    return `[addresses:${opts.addresses}][blocks:${fromBlock} - ${toBlock}]`;
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};
