import { WormchainRepository } from "../../repositories";
import { WormchainBlockLogs } from "../../entities/wormchain";
import { mapChain } from "../../../common/wormchain";
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
    opts: { addresses: string[]; chainId: number }
  ): Promise<WormchainBlockLogs[]> {
    const fromBlock = range.fromBlock;
    const toBlock = range.toBlock;
    const chain = mapChain(opts.chainId);

    const collectWormchainLogs: WormchainBlockLogs[] = [];

    if (fromBlock > toBlock) {
      this.logger.info(
        `[${chain}][exec] Invalid range [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
      );
      return [];
    }

    for (let blockNumber = fromBlock; blockNumber <= toBlock; blockNumber++) {
      const wormchainLogs = await this.blockRepo.getBlockLogs(
        opts.chainId,
        blockNumber,
        ATTRIBUTES_TYPES
      );

      if (wormchainLogs && wormchainLogs.transactions && wormchainLogs.transactions.length > 0) {
        collectWormchainLogs.push(wormchainLogs);
      }
    }

    this.logger.info(
      `[${chain}][exec] Got ${
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
